package domain

import "time"

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

type ErrorLocation struct {
	File      string
	Line      int
	Column    int
	Path      string
	Context   string
}

type ValidationError struct {
	Code      string
	Message   string
	Severity  Severity
	Location  *ErrorLocation
	Details   map[string]interface{}
	Timestamp time.Time
}

type ValidationReport struct {
	FilePath      string
	FileType      string
	IsValid       bool
	Errors        []ValidationError
	Warnings      []ValidationError
	Info          []ValidationError
	ValidationTime time.Time
	Duration      time.Duration
	Metadata      map[string]interface{}
}

func (r *ValidationReport) HasErrors() bool {
	return len(r.Errors) > 0
}

func (r *ValidationReport) HasWarnings() bool {
	return len(r.Warnings) > 0
}

func (r *ValidationReport) ErrorCount() int {
	return len(r.Errors)
}

func (r *ValidationReport) WarningCount() int {
	return len(r.Warnings)
}

func (r *ValidationReport) InfoCount() int {
	return len(r.Info)
}

func (r *ValidationReport) TotalIssues() int {
	return len(r.Errors) + len(r.Warnings) + len(r.Info)
}
