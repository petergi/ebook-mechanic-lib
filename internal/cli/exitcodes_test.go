package cli

import (
	"errors"
	"testing"

	"github.com/petergi/ebook-mechanic-lib/internal/domain"
)

func exitCode(t *testing.T, err error) int {
	t.Helper()

	var exitErr ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %v", err)
	}
	return exitErr.Code
}

func TestExitWithReport(t *testing.T) {
	ok := &domain.ValidationReport{}
	warn := &domain.ValidationReport{Warnings: []domain.ValidationError{{Code: "W"}}}
	errReport := &domain.ValidationReport{Errors: []domain.ValidationError{{Code: "E"}}}

	if code := exitCode(t, ExitWithReport(ok)); code != ExitCodeOK {
		t.Fatalf("expected OK exit code")
	}
	if code := exitCode(t, ExitWithReport(warn)); code != ExitCodeWarning {
		t.Fatalf("expected warning exit code")
	}
	if code := exitCode(t, ExitWithReport(errReport)); code != ExitCodeError {
		t.Fatalf("expected error exit code")
	}
}

func TestExitWithBatchResult(t *testing.T) {
	result := BatchResult{}
	if code := exitCode(t, ExitWithBatchResult(result)); code != ExitCodeOK {
		t.Fatalf("expected OK exit code")
	}

	result.HasWarnings = true
	if code := exitCode(t, ExitWithBatchResult(result)); code != ExitCodeWarning {
		t.Fatalf("expected warning exit code")
	}

	result.HasErrors = true
	if code := exitCode(t, ExitWithBatchResult(result)); code != ExitCodeError {
		t.Fatalf("expected error exit code")
	}
}
