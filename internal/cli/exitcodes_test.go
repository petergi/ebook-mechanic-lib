package cli

import (
	"testing"

	"github.com/example/project/internal/domain"
)

func TestExitWithReport(t *testing.T) {
	ok := &domain.ValidationReport{}
	warn := &domain.ValidationReport{Warnings: []domain.ValidationError{{Code: "W"}}}
	errReport := &domain.ValidationReport{Errors: []domain.ValidationError{{Code: "E"}}}

	if err := ExitWithReport(ok); err.(ExitError).Code != ExitCodeOK {
		t.Fatalf("expected OK exit code")
	}
	if err := ExitWithReport(warn); err.(ExitError).Code != ExitCodeWarning {
		t.Fatalf("expected warning exit code")
	}
	if err := ExitWithReport(errReport); err.(ExitError).Code != ExitCodeError {
		t.Fatalf("expected error exit code")
	}
}

func TestExitWithBatchResult(t *testing.T) {
	result := BatchResult{}
	if err := ExitWithBatchResult(result); err.(ExitError).Code != ExitCodeOK {
		t.Fatalf("expected OK exit code")
	}

	result.HasWarnings = true
	if err := ExitWithBatchResult(result); err.(ExitError).Code != ExitCodeWarning {
		t.Fatalf("expected warning exit code")
	}

	result.HasErrors = true
	if err := ExitWithBatchResult(result); err.(ExitError).Code != ExitCodeError {
		t.Fatalf("expected error exit code")
	}
}
