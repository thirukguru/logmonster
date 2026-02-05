package scanner

import (
	"context"
	"os"
	"path/filepath"

	"github.com/thiruk/logmonster/pkg/types"
)

// Walker handles directory walking with goroutine workers.
type Walker struct {
	config     Config
	fileChan   chan string
	resultChan chan types.FileInfo
}

// NewWalker creates a new directory walker.
func NewWalker(config Config) *Walker {
	return &Walker{
		config:     config,
		fileChan:   make(chan string, 1000),
		resultChan: make(chan types.FileInfo, 1000),
	}
}

// Walk walks all configured paths and returns file information.
func (w *Walker) Walk(ctx context.Context, paths []string) ([]types.FileInfo, error) {
	var files []types.FileInfo

	for _, basePath := range paths {
		err := filepath.WalkDir(basePath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				// Skip paths we can't access
				return nil
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Skip directories themselves
			if d.IsDir() {
				return nil
			}

			// Skip symlinks if configured
			if d.Type()&os.ModeSymlink != 0 && !w.config.FollowSymlinks {
				return nil
			}

			// Check exclude patterns
			if w.isExcluded(d.Name()) {
				return nil
			}

			info, err := d.Info()
			if err != nil {
				return nil
			}

			files = append(files, types.FileInfo{
				Path:       path,
				Size:       info.Size(),
				ModTime:    info.ModTime(),
				IsDir:      info.IsDir(),
				Permission: uint32(info.Mode().Perm()),
			})

			return nil
		})

		if err != nil && err != context.Canceled {
			return nil, err
		}
	}

	return files, nil
}

// isExcluded checks if a filename matches any exclude pattern.
func (w *Walker) isExcluded(name string) bool {
	for _, pattern := range w.config.ExcludePatterns {
		matched, _ := filepath.Match(pattern, name)
		if matched {
			return true
		}
	}
	return false
}
