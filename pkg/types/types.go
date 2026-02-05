package types

import "time"

// FileInfo represents information about a file during scanning.
type FileInfo struct {
	Path       string
	Size       int64
	ModTime    time.Time
	IsDir      bool
	Permission uint32
}

// FileGrowth represents the growth of a file between two snapshots.
type FileGrowth struct {
	Path        string
	InitialSize int64
	FinalSize   int64
	GrowthBytes int64
	GrowthRate  float64 // bytes per second
	Interval    time.Duration
}

// Snapshot represents a point-in-time snapshot of files.
type Snapshot struct {
	Timestamp time.Time
	Files     map[string]FileInfo
	TotalSize int64
	FileCount int
}

// ProcessInfo represents information about a process.
type ProcessInfo struct {
	PID        int32
	Name       string
	Cmdline    string
	Exe        string
	User       string
	StartTime  time.Time
	CPUPercent float64
	MemoryMB   float64
	WriteBytes int64
}

// ServiceInfo represents information about a systemd service.
type ServiceInfo struct {
	Unit        string
	Status      string
	MainPID     int32
	StartTime   time.Time
	Description string
}

// ScanResult represents the result of a scan operation.
type ScanResult struct {
	StartTime    time.Time
	EndTime      time.Time
	Interval     time.Duration
	Snapshot1    *Snapshot
	Snapshot2    *Snapshot
	GrowingFiles []FileGrowth
	TotalGrowth  int64
	Paths        []string
}

// SeverityLevel represents the severity of file growth.
type SeverityLevel int

const (
	SeverityLow    SeverityLevel = iota // < 1 MB/s
	SeverityMedium                      // < 10 MB/s
	SeverityHigh                        // >= 10 MB/s
)

// GetSeverity returns the severity level based on growth rate (bytes/sec).
func GetSeverity(bytesPerSec float64) SeverityLevel {
	mbPerSec := bytesPerSec / (1024 * 1024)
	switch {
	case mbPerSec >= 10:
		return SeverityHigh
	case mbPerSec >= 1:
		return SeverityMedium
	default:
		return SeverityLow
	}
}
