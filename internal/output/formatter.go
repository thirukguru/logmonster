package output

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Formatter handles formatted output to the terminal.
type Formatter struct {
	writer    io.Writer
	useColors bool
}

// NewFormatter creates a new formatter.
func NewFormatter(useColors bool) *Formatter {
	return &Formatter{
		writer:    os.Stdout,
		useColors: useColors,
	}
}

// SetWriter sets the output writer.
func (f *Formatter) SetWriter(w io.Writer) {
	f.writer = w
}

// Print prints a message.
func (f *Formatter) Print(msg string) {
	fmt.Fprint(f.writer, msg)
}

// Println prints a message with a newline.
func (f *Formatter) Println(msg string) {
	fmt.Fprintln(f.writer, msg)
}

// Printf prints a formatted message.
func (f *Formatter) Printf(format string, args ...interface{}) {
	fmt.Fprintf(f.writer, format, args...)
}

// Title prints a styled title.
func (f *Formatter) Title(title string) {
	if f.useColors {
		fmt.Fprintln(f.writer, TitleStyle.Render(title))
	} else {
		fmt.Fprintf(f.writer, "=== %s ===\n", title)
	}
}

// Success prints a success message.
func (f *Formatter) Success(msg string) {
	if f.useColors {
		fmt.Fprintln(f.writer, SuccessStyle.Render("✓ "+msg))
	} else {
		fmt.Fprintf(f.writer, "[OK] %s\n", msg)
	}
}

// Warning prints a warning message.
func (f *Formatter) Warning(msg string) {
	if f.useColors {
		fmt.Fprintln(f.writer, WarningStyle.Render("⚠️  "+msg))
	} else {
		fmt.Fprintf(f.writer, "[WARN] %s\n", msg)
	}
}

// Error prints an error message.
func (f *Formatter) Error(msg string) {
	if f.useColors {
		fmt.Fprintln(f.writer, ErrorStyle.Render("✗ "+msg))
	} else {
		fmt.Fprintf(f.writer, "[ERROR] %s\n", msg)
	}
}

// Info prints an info message.
func (f *Formatter) Info(msg string) {
	if f.useColors {
		fmt.Fprintln(f.writer, lipgloss.NewStyle().Foreground(ColorCyan).Render("→ "+msg))
	} else {
		fmt.Fprintf(f.writer, "[INFO] %s\n", msg)
	}
}

// Box prints content in a styled box.
func (f *Formatter) Box(title, content string) {
	if f.useColors {
		box := BoxStyle.Render(fmt.Sprintf("%s\n%s", title, content))
		fmt.Fprintln(f.writer, box)
	} else {
		fmt.Fprintf(f.writer, "+--- %s ---+\n%s\n+---+\n", title, content)
	}
}

// Header prints the application header for watch mode.
func (f *Formatter) Header(refresh int) {
	header := `╔════════════════════════════════════════════════════════════╗
║         LOG MONSTER DETECTOR - LIVE WATCH                  ║
║  Refresh: %ds | Press 'q' to quit                          ║
╚════════════════════════════════════════════════════════════╝`

	if f.useColors {
		styled := lipgloss.NewStyle().
			Foreground(ColorCyan).
			Bold(true).
			Render(fmt.Sprintf(header, refresh))
		fmt.Fprintln(f.writer, styled)
	} else {
		fmt.Fprintf(f.writer, header+"\n", refresh)
	}
}
