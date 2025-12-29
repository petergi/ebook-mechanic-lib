package domain

import "time"

// Severity represents the level of a validation finding.
type Severity string

const (
	// SeverityError indicates an error that makes a document invalid.
	SeverityError Severity = "error"
	// SeverityWarning indicates a warning that does not block validity.
	SeverityWarning Severity = "warning"
	// SeverityInfo indicates an informational finding.
	SeverityInfo Severity = "info"
)

// ErrorLocation describes where a validation error occurred.
type ErrorLocation struct {
	File    string
	Line    int
	Column  int
	Path    string
	Context string
}

// ValidationError represents a single validation finding.
type ValidationError struct {
	Code      string
	Message   string
	Severity  Severity
	Location  *ErrorLocation
	Details   map[string]interface{}
	Timestamp time.Time
}

// ValidationReport aggregates validation results for a file.
type ValidationReport struct {
	FilePath       string
	FileType       string
	IsValid        bool
	Errors         []ValidationError
	Warnings       []ValidationError
	Info           []ValidationError
	ValidationTime time.Time
	Duration       time.Duration
	Metadata       map[string]interface{}
}

// HasErrors reports whether any errors are present.
func (r *ValidationReport) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings reports whether any warnings are present.
func (r *ValidationReport) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// ErrorCount returns the number of errors.
func (r *ValidationReport) ErrorCount() int {
	return len(r.Errors)
}

// WarningCount returns the number of warnings.
func (r *ValidationReport) WarningCount() int {
	return len(r.Warnings)
}

// InfoCount returns the number of info messages.
func (r *ValidationReport) InfoCount() int {
	return len(r.Info)
}

// TotalIssues returns the total number of findings.
func (r *ValidationReport) TotalIssues() int {
	return len(r.Errors) + len(r.Warnings) + len(r.Info)
}
