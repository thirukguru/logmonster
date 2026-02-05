// Package scanner provides file scanning and growth detection functionality.
package scanner

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/thiruk/logmonster/pkg/types"
)

// Config holds scanner configuration.
type Config struct {
	Paths           []string
	Interval        time.Duration
	ThresholdBytes  int64
	WorkerCount     int
	MaxDepth        int
	FollowSymlinks  bool
	ExcludePatterns []string
}

// DefaultConfig returns a default scanner configuration.
func DefaultConfig() Config {
	return Config{
		Paths:          []string{"/var/log", "/tmp"},
		Interval:       5 * time.Second,
		ThresholdBytes: 10 * 1024 * 1024, // 10 MB
		WorkerCount:    4,
		MaxDepth:       10,
		FollowSymlinks: false,
	}
}

// Scanner handles file scanning and growth detection.
type Scanner struct {
	config Config
}

// New creates a new Scanner with the given configuration.
func New(config Config) *Scanner {
	if config.WorkerCount <= 0 {
		config.WorkerCount = 4
	}
	return &Scanner{config: config}
}

// Scan performs a full scan operation: takes two snapshots and calculates growth.
func (s *Scanner) Scan(ctx context.Context) (*types.ScanResult, error) {
	result := &types.ScanResult{
		StartTime: time.Now(),
		Paths:     s.config.Paths,
		Interval:  s.config.Interval,
	}

	// Take first snapshot
	snap1, err := s.TakeSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	result.Snapshot1 = snap1

	// Wait for interval
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(s.config.Interval):
	}

	// Take second snapshot
	snap2, err := s.TakeSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	result.Snapshot2 = snap2
	result.EndTime = time.Now()

	// Calculate growth
	result.GrowingFiles = s.CalculateGrowth(snap1, snap2)

	// Calculate total growth
	for _, g := range result.GrowingFiles {
		result.TotalGrowth += g.GrowthBytes
	}

	return result, nil
}

// TakeSnapshot takes a snapshot of all files in the configured paths.
func (s *Scanner) TakeSnapshot(ctx context.Context) (*types.Snapshot, error) {
	snapshot := &types.Snapshot{
		Timestamp: time.Now(),
		Files:     make(map[string]types.FileInfo),
	}

	fileChan := make(chan string, 1000)
	resultChan := make(chan types.FileInfo, 1000)
	errChan := make(chan error, 1)

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < s.config.WorkerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range fileChan {
				select {
				case <-ctx.Done():
					return
				default:
					info, err := s.statFile(path)
					if err != nil {
						// Skip files we can't stat (permission denied, deleted, etc.)
						continue
					}
					resultChan <- info
				}
			}
		}()
	}

	// Walk directories and feed file paths
	go func() {
		defer close(fileChan)
		for _, basePath := range s.config.Paths {
			select {
			case <-ctx.Done():
				return
			default:
				s.walkDirectory(ctx, basePath, fileChan, 0)
			}
		}
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results
	for info := range resultChan {
		snapshot.Files[info.Path] = info
		if !info.IsDir {
			snapshot.TotalSize += info.Size
			snapshot.FileCount++
		}
	}

	select {
	case err := <-errChan:
		return nil, err
	default:
	}

	return snapshot, nil
}

// walkDirectory recursively walks a directory and sends file paths to the channel.
func (s *Scanner) walkDirectory(ctx context.Context, path string, fileChan chan<- string, depth int) {
	if s.config.MaxDepth > 0 && depth > s.config.MaxDepth {
		return
	}

	select {
	case <-ctx.Done():
		return
	default:
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return // Skip directories we can't read
	}

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())

		// Skip symlinks if configured
		if entry.Type()&os.ModeSymlink != 0 && !s.config.FollowSymlinks {
			continue
		}

		// Check exclude patterns
		if s.isExcluded(entry.Name()) {
			continue
		}

		if entry.IsDir() {
			s.walkDirectory(ctx, fullPath, fileChan, depth+1)
		} else {
			select {
			case fileChan <- fullPath:
			case <-ctx.Done():
				return
			}
		}
	}
}

// isExcluded checks if a filename matches any exclude pattern.
func (s *Scanner) isExcluded(name string) bool {
	for _, pattern := range s.config.ExcludePatterns {
		matched, _ := filepath.Match(pattern, name)
		if matched {
			return true
		}
	}
	return false
}

// statFile returns file information for a path.
func (s *Scanner) statFile(path string) (types.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return types.FileInfo{}, err
	}

	return types.FileInfo{
		Path:       path,
		Size:       info.Size(),
		ModTime:    info.ModTime(),
		IsDir:      info.IsDir(),
		Permission: uint32(info.Mode().Perm()),
	}, nil
}

// CalculateGrowth calculates file growth between two snapshots.
func (s *Scanner) CalculateGrowth(snap1, snap2 *types.Snapshot) []types.FileGrowth {
	interval := snap2.Timestamp.Sub(snap1.Timestamp)
	if interval <= 0 {
		interval = time.Second // Prevent division by zero
	}

	var growing []types.FileGrowth

	for path, info2 := range snap2.Files {
		info1, exists := snap1.Files[path]
		if !exists {
			// New file - count entire size as growth
			if info2.Size >= s.config.ThresholdBytes {
				growing = append(growing, types.FileGrowth{
					Path:        path,
					InitialSize: 0,
					FinalSize:   info2.Size,
					GrowthBytes: info2.Size,
					GrowthRate:  float64(info2.Size) / interval.Seconds(),
					Interval:    interval,
				})
			}
			continue
		}

		growth := info2.Size - info1.Size
		if growth >= s.config.ThresholdBytes {
			growing = append(growing, types.FileGrowth{
				Path:        path,
				InitialSize: info1.Size,
				FinalSize:   info2.Size,
				GrowthBytes: growth,
				GrowthRate:  float64(growth) / interval.Seconds(),
				Interval:    interval,
			})
		}
	}

	// Sort by growth rate descending
	sort.Slice(growing, func(i, j int) bool {
		return growing[i].GrowthRate > growing[j].GrowthRate
	})

	return growing
}
