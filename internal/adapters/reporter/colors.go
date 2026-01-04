// Package reporter provides report formatting adapters.
package reporter

import (
	"fmt"

	"github.com/petergi/ebook-mechanic-lib/internal/domain"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGreen  = "\033[32m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// ColorScheme defines ANSI color mappings for output.
type ColorScheme struct {
	Error   string
	Warning string
	Info    string
	Success string
	Header  string
	Path    string
	Code    string
	Dim     string
	Reset   string
}

// NewColorScheme returns a color scheme configured for enabled output.
func NewColorScheme(enabled bool) *ColorScheme {
	if !enabled {
		return &ColorScheme{
			Error:   "",
			Warning: "",
			Info:    "",
			Success: "",
			Header:  "",
			Path:    "",
			Code:    "",
			Dim:     "",
			Reset:   "",
		}
	}

	return &ColorScheme{
		Error:   colorRed,
		Warning: colorYellow,
		Info:    colorBlue,
		Success: colorGreen,
		Header:  colorBold,
		Path:    colorCyan,
		Code:    colorDim,
		Dim:     colorDim,
		Reset:   colorReset,
	}
}

// Colorize wraps text in the specified ANSI color.
func (c *ColorScheme) Colorize(text, color string) string {
	if color == "" {
		return text
	}
	return fmt.Sprintf("%s%s%s", color, text, c.Reset)
}

// ColorizeError applies error styling.
func (c *ColorScheme) ColorizeError(text string) string {
	return c.Colorize(text, c.Error)
}

// ColorizeWarning applies warning styling.
func (c *ColorScheme) ColorizeWarning(text string) string {
	return c.Colorize(text, c.Warning)
}

// ColorizeInfo applies info styling.
func (c *ColorScheme) ColorizeInfo(text string) string {
	return c.Colorize(text, c.Info)
}

// ColorizeSuccess applies success styling.
func (c *ColorScheme) ColorizeSuccess(text string) string {
	return c.Colorize(text, c.Success)
}

// ColorizeHeader applies header styling.
func (c *ColorScheme) ColorizeHeader(text string) string {
	return c.Colorize(text, c.Header)
}

// ColorizePath applies path styling.
func (c *ColorScheme) ColorizePath(text string) string {
	return c.Colorize(text, c.Path)
}

// ColorizeCode applies code styling.
func (c *ColorScheme) ColorizeCode(text string) string {
	return c.Colorize(text, c.Code)
}

// ColorizeDim applies dim styling.
func (c *ColorScheme) ColorizeDim(text string) string {
	return c.Colorize(text, c.Dim)
}

// ColorizeForSeverity applies styling based on severity.
func (c *ColorScheme) ColorizeForSeverity(text string, severity domain.Severity) string {
	switch severity {
	case domain.SeverityError:
		return c.ColorizeError(text)
	case domain.SeverityWarning:
		return c.ColorizeWarning(text)
	case domain.SeverityInfo:
		return c.ColorizeInfo(text)
	default:
		return text
	}
}
