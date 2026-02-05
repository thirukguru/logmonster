package util

import "fmt"

// FormatBytes formats bytes into human-readable format.
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatRate formats bytes per second into human-readable format.
func FormatRate(bytesPerSec float64) string {
	return FormatBytes(int64(bytesPerSec)) + "/s"
}

// FormatBytesWithSign formats bytes with a +/- sign.
func FormatBytesWithSign(bytes int64) string {
	if bytes >= 0 {
		return "+" + FormatBytes(bytes)
	}
	return "-" + FormatBytes(-bytes)
}
