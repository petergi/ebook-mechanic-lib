package reporter

import "github.com/example/project/internal/domain"

type Filter struct {
	Severities []domain.Severity
	Categories []string
	Standards  []string
	MinSeverity domain.Severity
}

func NewFilter() *Filter {
	return &Filter{
		Severities: make([]domain.Severity, 0),
		Categories: make([]string, 0),
		Standards:  make([]string, 0),
	}
}

func (f *Filter) Matches(err domain.ValidationError) bool {
	if len(f.Severities) > 0 {
		found := false
		for _, sev := range f.Severities {
			if err.Severity == sev {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if f.MinSeverity != "" {
		if !f.meetsMinSeverity(err.Severity) {
			return false
		}
	}

	if len(f.Categories) > 0 {
		if err.Details == nil {
			return false
		}
		category, ok := err.Details["category"].(string)
		if !ok {
			return false
		}
		found := false
		for _, cat := range f.Categories {
			if category == cat {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(f.Standards) > 0 {
		if err.Details == nil {
			return false
		}
		standard, ok := err.Details["standard"].(string)
		if !ok {
			return false
		}
		found := false
		for _, std := range f.Standards {
			if standard == std {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (f *Filter) meetsMinSeverity(severity domain.Severity) bool {
	severityOrder := map[domain.Severity]int{
		domain.SeverityInfo:    1,
		domain.SeverityWarning: 2,
		domain.SeverityError:   3,
	}

	minLevel := severityOrder[f.MinSeverity]
	currentLevel := severityOrder[severity]

	return currentLevel >= minLevel
}

func (f *Filter) FilterErrors(errors []domain.ValidationError) []domain.ValidationError {
	if f == nil {
		return errors
	}

	filtered := make([]domain.ValidationError, 0)
	for _, err := range errors {
		if f.Matches(err) {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

func FilterReportBySeverity(report *domain.ValidationReport, severities []domain.Severity) *domain.ValidationReport {
	filtered := &domain.ValidationReport{
		FilePath:       report.FilePath,
		FileType:       report.FileType,
		IsValid:        report.IsValid,
		ValidationTime: report.ValidationTime,
		Duration:       report.Duration,
		Metadata:       report.Metadata,
		Errors:         make([]domain.ValidationError, 0),
		Warnings:       make([]domain.ValidationError, 0),
		Info:           make([]domain.ValidationError, 0),
	}

	filter := &Filter{Severities: severities}

	filtered.Errors = filter.FilterErrors(report.Errors)
	filtered.Warnings = filter.FilterErrors(report.Warnings)
	filtered.Info = filter.FilterErrors(report.Info)

	return filtered
}
