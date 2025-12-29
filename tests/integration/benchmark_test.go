package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/example/project/internal/adapters/epub"
	"github.com/example/project/internal/adapters/pdf"
)

func BenchmarkEPUBValidation_Minimal(b *testing.B) {
	validator := epub.NewEPUBValidator()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "epub", "valid", "minimal.epub")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateFile(ctx, testFile)
		if err != nil {
			b.Fatalf("ValidateFile failed: %v", err)
		}
	}
}

func BenchmarkEPUBValidation_Large100Chapters(b *testing.B) {
	validator := epub.NewEPUBValidator()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "epub", "valid", "large_100_chapters.epub")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateFile(ctx, testFile)
		if err != nil {
			b.Fatalf("ValidateFile failed: %v", err)
		}
	}
}

func BenchmarkPDFValidation_Minimal(b *testing.B) {
	validator := pdf.NewStructureValidator()

	testFile := filepath.Join("..", "..", "testdata", "pdf", "valid", "minimal.pdf")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateFile(testFile)
		if err != nil {
			b.Fatalf("ValidateFile failed: %v", err)
		}
	}
}

func BenchmarkPDFValidation_Large100Pages(b *testing.B) {
	validator := pdf.NewStructureValidator()

	testFile := filepath.Join("..", "..", "testdata", "pdf", "valid", "large_100_pages.pdf")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateFile(testFile)
		if err != nil {
			b.Fatalf("ValidateFile failed: %v", err)
		}
	}
}
