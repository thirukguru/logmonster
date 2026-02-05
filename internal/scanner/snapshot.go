package scanner

import (
	"encoding/json"
	"os"
	"time"

	"github.com/thiruk/logmonster/pkg/types"
)

// SnapshotStore handles saving and loading snapshots to disk.
type SnapshotStore struct {
	basePath string
}

// NewSnapshotStore creates a new snapshot store.
func NewSnapshotStore(basePath string) *SnapshotStore {
	return &SnapshotStore{basePath: basePath}
}

// Save saves a snapshot to disk.
func (s *SnapshotStore) Save(snapshot *types.Snapshot, filename string) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// Load loads a snapshot from disk.
func (s *SnapshotStore) Load(filename string) (*types.Snapshot, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var snapshot types.Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

// CompareSnapshots compares two snapshots and returns the differences.
func CompareSnapshots(snap1, snap2 *types.Snapshot, thresholdBytes int64) []types.FileGrowth {
	interval := snap2.Timestamp.Sub(snap1.Timestamp)
	if interval <= 0 {
		interval = time.Second
	}

	var growing []types.FileGrowth

	// Find files that grew
	for path, info2 := range snap2.Files {
		if info2.IsDir {
			continue
		}

		info1, exists := snap1.Files[path]
		if !exists {
			// New file
			if info2.Size >= thresholdBytes {
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
		if growth >= thresholdBytes {
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

	return growing
}
