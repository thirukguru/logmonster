package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/thiruk/logmonster/pkg/types"
	"github.com/thiruk/logmonster/pkg/util"
)

// Styles for table rendering.
var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorGray)

	CellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(lipgloss.Color("#5A56E0")).
			Padding(0, 2).
			MarginBottom(1)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Padding(1, 2)

	WarningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorYellow)

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorRed)

	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGreen)
)

// Table represents a formatted table for terminal output.
type Table struct {
	headers []string
	rows    [][]string
	widths  []int
}

// NewTable creates a new table with the given headers.
func NewTable(headers ...string) *Table {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	return &Table{
		headers: headers,
		widths:  widths,
	}
}

// AddRow adds a row to the table.
func (t *Table) AddRow(cells ...string) {
	// Pad or truncate to match header count
	row := make([]string, len(t.headers))
	for i := range row {
		if i < len(cells) {
			row[i] = cells[i]
			if len(cells[i]) > t.widths[i] {
				t.widths[i] = len(cells[i])
			}
		}
	}
	t.rows = append(t.rows, row)
}

// Render renders the table as a string.
func (t *Table) Render() string {
	var sb strings.Builder

	// Top border
	sb.WriteString("┌")
	for i, w := range t.widths {
		sb.WriteString(strings.Repeat("─", w+2))
		if i < len(t.widths)-1 {
			sb.WriteString("┬")
		}
	}
	sb.WriteString("┐\n")

	// Header row
	sb.WriteString("│")
	for i, h := range t.headers {
		sb.WriteString(" ")
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render(padRight(h, t.widths[i])))
		sb.WriteString(" │")
	}
	sb.WriteString("\n")

	// Header separator
	sb.WriteString("├")
	for i, w := range t.widths {
		sb.WriteString(strings.Repeat("─", w+2))
		if i < len(t.widths)-1 {
			sb.WriteString("┼")
		}
	}
	sb.WriteString("┤\n")

	// Data rows
	for _, row := range t.rows {
		sb.WriteString("│")
		for i, cell := range row {
			sb.WriteString(" ")
			sb.WriteString(padRight(cell, t.widths[i]))
			sb.WriteString(" │")
		}
		sb.WriteString("\n")
	}

	// Bottom border
	sb.WriteString("└")
	for i, w := range t.widths {
		sb.WriteString(strings.Repeat("─", w+2))
		if i < len(t.widths)-1 {
			sb.WriteString("┴")
		}
	}
	sb.WriteString("┘")

	return sb.String()
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// RenderGrowthTable renders a table of file growth information.
func RenderGrowthTable(files []types.FileGrowth) string {
	table := NewTable("FILE", "GROWTH", "GROWTH/SEC")

	for _, f := range files {
		emoji := GetSeverityEmoji(f.GrowthRate)
		table.AddRow(
			truncatePath(f.Path, 40),
			util.FormatBytesWithSign(f.GrowthBytes),
			fmt.Sprintf("%s %s", emoji, util.FormatRate(f.GrowthRate)),
		)
	}

	return table.Render()
}

// RenderProcessInfo renders process information in a box.
func RenderProcessInfo(info types.ProcessInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("│ PID:          %-40d │\n", info.PID))
	sb.WriteString(fmt.Sprintf("│ Process:      %-40s │\n", info.Name))
	sb.WriteString(fmt.Sprintf("│ Command:      %-40s │\n", truncate(info.Cmdline, 40)))
	sb.WriteString(fmt.Sprintf("│ User:         %-40s │\n", info.User))
	sb.WriteString(fmt.Sprintf("│ Started:      %-40s │\n", info.StartTime.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("│ CPU:          %-40.1f%% │\n", info.CPUPercent))
	sb.WriteString(fmt.Sprintf("│ Memory:       %-40s │\n", fmt.Sprintf("%.1f MB", info.MemoryMB)))

	return BoxStyle.Render(sb.String())
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
