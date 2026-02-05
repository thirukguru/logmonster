// Package output provides terminal output formatting utilities.
package output

import (
	"github.com/charmbracelet/lipgloss"
)

// Color definitions for severity levels.
var (
	ColorGreen  = lipgloss.Color("#00FF00")
	ColorYellow = lipgloss.Color("#FFFF00")
	ColorRed    = lipgloss.Color("#FF0000")
	ColorCyan   = lipgloss.Color("#00FFFF")
	ColorWhite  = lipgloss.Color("#FFFFFF")
	ColorGray   = lipgloss.Color("#808080")
)

// Emoji indicators for severity.
const (
	EmojiGreen  = "🟢"
	EmojiYellow = "🟡"
	EmojiRed    = "🔴"
)

// GetSeverityColor returns the color for a given write rate (bytes/sec).
func GetSeverityColor(bytesPerSec float64) lipgloss.Color {
	mbPerSec := bytesPerSec / (1024 * 1024)
	switch {
	case mbPerSec >= 10:
		return ColorRed
	case mbPerSec >= 1:
		return ColorYellow
	default:
		return ColorGreen
	}
}

// GetSeverityEmoji returns the emoji for a given write rate (bytes/sec).
func GetSeverityEmoji(bytesPerSec float64) string {
	mbPerSec := bytesPerSec / (1024 * 1024)
	switch {
	case mbPerSec >= 10:
		return EmojiRed
	case mbPerSec >= 1:
		return EmojiYellow
	default:
		return EmojiGreen
	}
}
