package reporter

import (
	"testing"
	"time"

	"github.com/petergi/ebook-mechanic-lib/internal/domain"
)

func TestFilter_Matches(t *testing.T) {
	tests := []struct {
		name   string
		filter *Filter
		err    domain.ValidationError
		want   bool
	}{
		{
			name:   "empty filter matches everything",
			filter: NewFilter(),
			err: domain.ValidationError{
				Code:     "TEST-001",
				Severity: domain.SeverityError,
			},
			want: true,
		},
		{
			name: "severity filter matches",
			filter: &Filter{
				Severities: []domain.Severity{domain.SeverityError},
			},
			err: domain.ValidationError{
				Code:     "TEST-001",
				Severity: domain.SeverityError,
			},
			want: true,
		},
		{
			name: "severity filter does not match",
			filter: &Filter{
				Severities: []domain.Severity{domain.SeverityError},
			},
			err: domain.ValidationError{
				Code:     "TEST-001",
				Severity: domain.SeverityWarning,
			},
			want: false,
		},
		{
			name: "multiple severity filter matches",
			filter: &Filter{
				Severities: []domain.Severity{domain.SeverityError, domain.SeverityWarning},
			},
			err: domain.ValidationError{
				Code:     "TEST-001",
				Severity: domain.SeverityWarning,
			},
			want: true,
		},
		{
			name: "min severity matches error",
			filter: &Filter{
				MinSeverity: domain.SeverityWarning,
			},
			err: domain.ValidationError{
				Code:     "TEST-001",
				Severity: domain.SeverityError,
			},
			want: true,
		},
		{
			name: "min severity filters info",
			filter: &Filter{
				MinSeverity: domain.SeverityWarning,
			},
			err: domain.ValidationError{
				Code:     "TEST-001",
				Severity: domain.SeverityInfo,
			},
			want: false,
		},
		{
			name: "category filter matches",
			filter: &Filter{
				Categories: []string{"structure"},
			},
			err: domain.ValidationError{
				Code:     "TEST-001",
				Severity: domain.SeverityError,
				Details: map[string]interface{}{
					"category": "structure",
				},
			},
			want: true,
		},
		{
			name: "category filter does not match",
			filter: &Filter{
				Categories: []string{"structure"},
			},
			err: domain.ValidationError{
				Code:     "TEST-001",
				Severity: domain.SeverityError,
				Details: map[string]interface{}{
					"category": "metadata",
				},
			},
			want: false,
		},
		{
			name: "category filter no details",
			filter: &Filter{
				Categories: []string{"structure"},
			},
			err: domain.ValidationError{
				Code:     "TEST-001",
				Severity: domain.SeverityError,
			},
			want: false,
		},
		{
			name: "standard filter matches",
			filter: &Filter{
				Standards: []string{"EPUB3"},
			},
			err: domain.ValidationError{
				Code:     "TEST-001",
				Severity: domain.SeverityError,
				Details: map[string]interface{}{
					"standard": "EPUB3",
				},
			},
			want: true,
		},
		{
			name: "standard filter does not match",
			filter: &Filter{
				Standards: []string{"EPUB3"},
			},
			err: domain.ValidationError{
				Code:     "TEST-001",
				Severity: domain.SeverityError,
				Details: map[string]interface{}{
					"standard": "PDF/A",
				},
			},
			want: false,
		},
		{
			name: "combined filters all match",
			filter: &Filter{
				Severities: []domain.Severity{domain.SeverityError},
				Categories: []string{"structure"},
				Standards:  []string{"EPUB3"},
			},
			err: domain.ValidationError{
				Code:     "TEST-001",
				Severity: domain.SeverityError,
				Details: map[string]interface{}{
					"category": "structure",
					"standard": "EPUB3",
				},
			},
			want: true,
		},
		{
			name: "combined filters severity does not match",
			filter: &Filter{
				Severities: []domain.Severity{domain.SeverityError},
				Categories: []string{"structure"},
			},
			err: domain.ValidationError{
				Code:     "TEST-001",
				Severity: domain.SeverityWarning,
				Details: map[string]interface{}{
					"category": "structure",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Matches(tt.err)
			if got != tt.want {
				t.Errorf("Filter.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_FilterErrors(t *testing.T) {
	errors := []domain.ValidationError{
		{
			Code:     "TEST-001",
			Severity: domain.SeverityError,
			Details: map[string]interface{}{
				"category": "structure",
			},
		},
		{
			Code:     "TEST-002",
			Severity: domain.SeverityWarning,
			Details: map[string]interface{}{
				"category": "metadata",
			},
		},
		{
			Code:     "TEST-003",
			Severity: domain.SeverityInfo,
			Details: map[string]interface{}{
				"category": "structure",
			},
		},
	}

	tests := []struct {
		name   string
		filter *Filter
		want   int
	}{
		{
			name:   "no filter returns all",
			filter: nil,
			want:   3,
		},
		{
			name:   "empty filter returns all",
			filter: NewFilter(),
			want:   3,
		},
		{
			name: "severity filter",
			filter: &Filter{
				Severities: []domain.Severity{domain.SeverityError},
			},
			want: 1,
		},
		{
			name: "category filter",
			filter: &Filter{
				Categories: []string{"structure"},
			},
			want: 2,
		},
		{
			name: "min severity filter",
			filter: &Filter{
				MinSeverity: domain.SeverityWarning,
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.FilterErrors(errors)
			if len(got) != tt.want {
				t.Errorf("Filter.FilterErrors() returned %d errors, want %d", len(got), tt.want)
			}
		})
	}
}

func TestFilterReportBySeverity(t *testing.T) {
	report := &domain.ValidationReport{
		FilePath: "test.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{Code: "E1", Severity: domain.SeverityError},
			{Code: "E2", Severity: domain.SeverityError},
		},
		Warnings: []domain.ValidationError{
			{Code: "W1", Severity: domain.SeverityWarning},
		},
		Info: []domain.ValidationError{
			{Code: "I1", Severity: domain.SeverityInfo},
		},
		ValidationTime: time.Now(),
	}

	filtered := FilterReportBySeverity(report, []domain.Severity{domain.SeverityError})

	if len(filtered.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(filtered.Errors))
	}
	if len(filtered.Warnings) != 0 {
		t.Errorf("Expected 0 warnings, got %d", len(filtered.Warnings))
	}
	if len(filtered.Info) != 0 {
		t.Errorf("Expected 0 info, got %d", len(filtered.Info))
	}
}
