package reporter

import (
	"fmt"

	"github.com/example/project/internal/domain"
)

const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorGreen   = "\033[32m"
	colorCyan    = "\033[36m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
)

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

func (c *ColorScheme) Colorize(text, color string) string {
	if color == "" {
		return text
	}
	return fmt.Sprintf("%s%s%s", color, text, c.Reset)
}

func (c *ColorScheme) ColorizeError(text string) string {
	return c.Colorize(text, c.Error)
}

func (c *ColorScheme) ColorizeWarning(text string) string {
	return c.Colorize(text, c.Warning)
}

func (c *ColorScheme) ColorizeInfo(text string) string {
	return c.Colorize(text, c.Info)
}

func (c *ColorScheme) ColorizeSuccess(text string) string {
	return c.Colorize(text, c.Success)
}

func (c *ColorScheme) ColorizeHeader(text string) string {
	return c.Colorize(text, c.Header)
}

func (c *ColorScheme) ColorizePath(text string) string {
	return c.Colorize(text, c.Path)
}

func (c *ColorScheme) ColorizeCode(text string) string {
	return c.Colorize(text, c.Code)
}

func (c *ColorScheme) ColorizeDim(text string) string {
	return c.Colorize(text, c.Dim)
}

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
