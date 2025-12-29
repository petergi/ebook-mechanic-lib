package reporter

import (
	"strings"
	"testing"

	"github.com/example/project/internal/domain"
)

func TestNewColorScheme_Enabled(t *testing.T) {
	scheme := NewColorScheme(true)

	if scheme.Error == "" {
		t.Error("Expected error color to be set when colors enabled")
	}

	if scheme.Warning == "" {
		t.Error("Expected warning color to be set when colors enabled")
	}

	if scheme.Info == "" {
		t.Error("Expected info color to be set when colors enabled")
	}

	if scheme.Success == "" {
		t.Error("Expected success color to be set when colors enabled")
	}

	if scheme.Reset == "" {
		t.Error("Expected reset color to be set when colors enabled")
	}
}

func TestNewColorScheme_Disabled(t *testing.T) {
	scheme := NewColorScheme(false)

	if scheme.Error != "" {
		t.Error("Expected error color to be empty when colors disabled")
	}

	if scheme.Warning != "" {
		t.Error("Expected warning color to be empty when colors disabled")
	}

	if scheme.Info != "" {
		t.Error("Expected info color to be empty when colors disabled")
	}

	if scheme.Success != "" {
		t.Error("Expected success color to be empty when colors disabled")
	}

	if scheme.Reset != "" {
		t.Error("Expected reset color to be empty when colors disabled")
	}
}

func TestColorScheme_Colorize_Enabled(t *testing.T) {
	scheme := NewColorScheme(true)
	text := "test message"

	result := scheme.Colorize(text, colorRed)

	if !strings.Contains(result, colorRed) {
		t.Error("Expected result to contain color code")
	}

	if !strings.Contains(result, text) {
		t.Error("Expected result to contain original text")
	}

	if !strings.Contains(result, colorReset) {
		t.Error("Expected result to contain reset code")
	}

	expected := colorRed + text + colorReset
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestColorScheme_Colorize_Disabled(t *testing.T) {
	scheme := NewColorScheme(false)
	text := "test message"

	result := scheme.Colorize(text, "")

	if result != text {
		t.Errorf("Expected '%s', got '%s'", text, result)
	}

	if strings.Contains(result, "\033[") {
		t.Error("Did not expect any color codes when disabled")
	}
}

func TestColorScheme_ColorizeError(t *testing.T) {
	scheme := NewColorScheme(true)
	text := "error message"

	result := scheme.ColorizeError(text)

	if !strings.Contains(result, colorRed) {
		t.Error("Expected red color for error")
	}
}

func TestColorScheme_ColorizeWarning(t *testing.T) {
	scheme := NewColorScheme(true)
	text := "warning message"

	result := scheme.ColorizeWarning(text)

	if !strings.Contains(result, colorYellow) {
		t.Error("Expected yellow color for warning")
	}
}

func TestColorScheme_ColorizeInfo(t *testing.T) {
	scheme := NewColorScheme(true)
	text := "info message"

	result := scheme.ColorizeInfo(text)

	if !strings.Contains(result, colorBlue) {
		t.Error("Expected blue color for info")
	}
}

func TestColorScheme_ColorizeSuccess(t *testing.T) {
	scheme := NewColorScheme(true)
	text := "success message"

	result := scheme.ColorizeSuccess(text)

	if !strings.Contains(result, colorGreen) {
		t.Error("Expected green color for success")
	}
}

func TestColorScheme_ColorizeHeader(t *testing.T) {
	scheme := NewColorScheme(true)
	text := "header text"

	result := scheme.ColorizeHeader(text)

	if !strings.Contains(result, colorBold) {
		t.Error("Expected bold for header")
	}
}

func TestColorScheme_ColorizePath(t *testing.T) {
	scheme := NewColorScheme(true)
	text := "/path/to/file"

	result := scheme.ColorizePath(text)

	if !strings.Contains(result, colorCyan) {
		t.Error("Expected cyan color for path")
	}
}

func TestColorScheme_ColorizeCode(t *testing.T) {
	scheme := NewColorScheme(true)
	text := "code block"

	result := scheme.ColorizeCode(text)

	if !strings.Contains(result, colorDim) {
		t.Error("Expected dim color for code")
	}
}

func TestColorScheme_ColorizeDim(t *testing.T) {
	scheme := NewColorScheme(true)
	text := "dim text"

	result := scheme.ColorizeDim(text)

	if !strings.Contains(result, colorDim) {
		t.Error("Expected dim color")
	}
}

func TestColorScheme_ColorizeForSeverity(t *testing.T) {
	scheme := NewColorScheme(true)

	tests := []struct {
		name     string
		severity domain.Severity
		expected string
	}{
		{
			name:     "error severity",
			severity: domain.SeverityError,
			expected: colorRed,
		},
		{
			name:     "warning severity",
			severity: domain.SeverityWarning,
			expected: colorYellow,
		},
		{
			name:     "info severity",
			severity: domain.SeverityInfo,
			expected: colorBlue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := "test message"
			result := scheme.ColorizeForSeverity(text, tt.severity)

			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected color code '%s' for severity '%s'", tt.expected, tt.severity)
			}

			if !strings.Contains(result, text) {
				t.Error("Expected result to contain original text")
			}
		})
	}
}

func TestColorScheme_ColorizeForSeverity_Unknown(t *testing.T) {
	scheme := NewColorScheme(true)
	text := "test message"

	result := scheme.ColorizeForSeverity(text, domain.Severity("unknown"))

	if result != text {
		t.Error("Expected original text for unknown severity")
	}
}

func TestColorScheme_AllColors_Disabled(t *testing.T) {
	scheme := NewColorScheme(false)
	text := "test"

	testCases := []struct {
		name   string
		result string
	}{
		{"ColorizeError", scheme.ColorizeError(text)},
		{"ColorizeWarning", scheme.ColorizeWarning(text)},
		{"ColorizeInfo", scheme.ColorizeInfo(text)},
		{"ColorizeSuccess", scheme.ColorizeSuccess(text)},
		{"ColorizeHeader", scheme.ColorizeHeader(text)},
		{"ColorizePath", scheme.ColorizePath(text)},
		{"ColorizeCode", scheme.ColorizeCode(text)},
		{"ColorizeDim", scheme.ColorizeDim(text)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.result != text {
				t.Errorf("%s: Expected '%s', got '%s'", tc.name, text, tc.result)
			}
			if strings.Contains(tc.result, "\033[") {
				t.Errorf("%s: Did not expect color codes when disabled", tc.name)
			}
		})
	}
}
