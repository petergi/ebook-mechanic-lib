package cli

import "github.com/example/project/internal/domain"

type RepairOptions struct {
	OutputPath string
	InPlace    bool
	Backup     bool
	BackupDir  string
}

type BatchOptions struct {
	Workers     int
	QueueSize   int
	MaxDepth    int
	Extensions  []string
	Ignore      []string
	Progress    string
	SummaryOnly bool
	OutputPath  string
	Repair      RepairOptions
}

type BatchResult struct {
	Reports       []*domain.ValidationReport
	Total         int
	Processed     int
	Skipped       int
	Failed        int
	HasWarnings   bool
	HasErrors     bool
	InternalError error
}
