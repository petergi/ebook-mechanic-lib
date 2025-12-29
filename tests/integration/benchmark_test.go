// Package integration provides comprehensive benchmark tests for performance monitoring
// and regression detection across EPUB/PDF validation, reporter formatting, and repair operations.
//
// Benchmark Categories:
// - EPUB Validation: Small (<1MB), Medium (1-10MB), Large (>10MB)
// - PDF Validation: Small, Medium, Large file sizes
// - Reporter Formatting: Various error set sizes (10, 100, 1000, 10000 errors)
// - Repair Service: Preview and Apply operations
//
// Run benchmarks: make test-bench
// Compare benchmarks: go test -bench=. -benchmem -count=5 > new.txt
package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/project/internal/adapters/epub"
	"github.com/example/project/internal/adapters/pdf"
	"github.com/example/project/internal/adapters/reporter"
	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
)

// EPUB Validation Benchmarks - Various File Sizes

func BenchmarkEPUBValidation_Small_Minimal(b *testing.B) {
	validator := epub.NewEPUBValidator()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "epub", "valid", "minimal.epub")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateFile(ctx, testFile)
		if err != nil {
			b.Fatalf("ValidateFile failed: %v", err)
		}
	}
}

func BenchmarkEPUBValidation_Medium_100Chapters(b *testing.B) {
	validator := epub.NewEPUBValidator()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "epub", "valid", "large_100_chapters.epub")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateFile(ctx, testFile)
		if err != nil {
			b.Fatalf("ValidateFile failed: %v", err)
		}
	}
}

func BenchmarkEPUBValidation_Large_500Chapters(b *testing.B) {
	validator := epub.NewEPUBValidator()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "epub", "valid", "large_500_chapters.epub")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateFile(ctx, testFile)
		if err != nil {
			b.Fatalf("ValidateFile failed: %v", err)
		}
	}
}

func BenchmarkEPUBValidation_Structure_Small(b *testing.B) {
	validator := epub.NewEPUBValidator()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "epub", "valid", "minimal.epub")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateStructure(ctx, testFile)
		if err != nil {
			b.Fatalf("ValidateStructure failed: %v", err)
		}
	}
}

func BenchmarkEPUBValidation_Metadata_Small(b *testing.B) {
	validator := epub.NewEPUBValidator()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "epub", "valid", "minimal.epub")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateMetadata(ctx, testFile)
		if err != nil {
			b.Fatalf("ValidateMetadata failed: %v", err)
		}
	}
}

func BenchmarkEPUBValidation_Content_Medium(b *testing.B) {
	validator := epub.NewEPUBValidator()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "epub", "valid", "large_100_chapters.epub")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateContent(ctx, testFile)
		if err != nil {
			b.Fatalf("ValidateContent failed: %v", err)
		}
	}
}

// PDF Validation Benchmarks - Various File Sizes

func BenchmarkPDFValidation_Small_Minimal(b *testing.B) {
	validator := pdf.NewStructureValidator()

	testFile := filepath.Join("..", "..", "testdata", "pdf", "valid", "minimal.pdf")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateFile(testFile)
		if err != nil {
			b.Fatalf("ValidateFile failed: %v", err)
		}
	}
}

func BenchmarkPDFValidation_Medium_100Pages(b *testing.B) {
	validator := pdf.NewStructureValidator()

	testFile := filepath.Join("..", "..", "testdata", "pdf", "valid", "large_100_pages.pdf")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateFile(testFile)
		if err != nil {
			b.Fatalf("ValidateFile failed: %v", err)
		}
	}
}

func BenchmarkPDFValidation_Large_500Pages(b *testing.B) {
	validator := pdf.NewStructureValidator()

	testFile := filepath.Join("..", "..", "testdata", "pdf", "valid", "large_500_pages.pdf")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateFile(testFile)
		if err != nil {
			b.Fatalf("ValidateFile failed: %v", err)
		}
	}
}

func BenchmarkPDFValidation_Reader_Small(b *testing.B) {
	validator := pdf.NewStructureValidator()

	testFile := filepath.Join("..", "..", "testdata", "pdf", "valid", "minimal.pdf")
	data, err := os.ReadFile(testFile)
	if err != nil {
		b.Skipf("Test file not found: %s", testFile)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateBytes(data)
		if err != nil {
			b.Fatalf("ValidateBytes failed: %v", err)
		}
	}
}

// Reporter Formatting Benchmarks - Large Error Sets

func BenchmarkReporter_JSON_SmallErrorSet(b *testing.B) {
	ctx := context.Background()
	rep := reporter.NewJSONReporter()
	report := createBenchmarkReport("test.epub", 10)
	options := &ports.ReportOptions{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := rep.Format(ctx, report, options)
		if err != nil {
			b.Fatalf("Format failed: %v", err)
		}
	}
}

func BenchmarkReporter_JSON_MediumErrorSet(b *testing.B) {
	ctx := context.Background()
	rep := reporter.NewJSONReporter()
	report := createBenchmarkReport("test.epub", 100)
	options := &ports.ReportOptions{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := rep.Format(ctx, report, options)
		if err != nil {
			b.Fatalf("Format failed: %v", err)
		}
	}
}

func BenchmarkReporter_JSON_LargeErrorSet(b *testing.B) {
	ctx := context.Background()
	rep := reporter.NewJSONReporter()
	report := createBenchmarkReport("test.epub", 1000)
	options := &ports.ReportOptions{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := rep.Format(ctx, report, options)
		if err != nil {
			b.Fatalf("Format failed: %v", err)
		}
	}
}

func BenchmarkReporter_JSON_VeryLargeErrorSet(b *testing.B) {
	ctx := context.Background()
	rep := reporter.NewJSONReporter()
	report := createBenchmarkReport("test.epub", 10000)
	options := &ports.ReportOptions{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := rep.Format(ctx, report, options)
		if err != nil {
			b.Fatalf("Format failed: %v", err)
		}
	}
}

func BenchmarkReporter_Markdown_SmallErrorSet(b *testing.B) {
	ctx := context.Background()
	rep := reporter.NewMarkdownReporter()
	report := createBenchmarkReport("test.epub", 10)
	options := &ports.ReportOptions{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := rep.Format(ctx, report, options)
		if err != nil {
			b.Fatalf("Format failed: %v", err)
		}
	}
}

func BenchmarkReporter_Markdown_MediumErrorSet(b *testing.B) {
	ctx := context.Background()
	rep := reporter.NewMarkdownReporter()
	report := createBenchmarkReport("test.epub", 100)
	options := &ports.ReportOptions{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := rep.Format(ctx, report, options)
		if err != nil {
			b.Fatalf("Format failed: %v", err)
		}
	}
}

func BenchmarkReporter_Markdown_LargeErrorSet(b *testing.B) {
	ctx := context.Background()
	rep := reporter.NewMarkdownReporter()
	report := createBenchmarkReport("test.epub", 1000)
	options := &ports.ReportOptions{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := rep.Format(ctx, report, options)
		if err != nil {
			b.Fatalf("Format failed: %v", err)
		}
	}
}

func BenchmarkReporter_Text_SmallErrorSet(b *testing.B) {
	ctx := context.Background()
	rep := reporter.NewTextReporter()
	report := createBenchmarkReport("test.epub", 10)
	options := &ports.ReportOptions{ColorEnabled: false}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := rep.Format(ctx, report, options)
		if err != nil {
			b.Fatalf("Format failed: %v", err)
		}
	}
}

func BenchmarkReporter_Text_MediumErrorSet(b *testing.B) {
	ctx := context.Background()
	rep := reporter.NewTextReporter()
	report := createBenchmarkReport("test.epub", 100)
	options := &ports.ReportOptions{ColorEnabled: false}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := rep.Format(ctx, report, options)
		if err != nil {
			b.Fatalf("Format failed: %v", err)
		}
	}
}

func BenchmarkReporter_Text_LargeErrorSet(b *testing.B) {
	ctx := context.Background()
	rep := reporter.NewTextReporter()
	report := createBenchmarkReport("test.epub", 1000)
	options := &ports.ReportOptions{ColorEnabled: false}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := rep.Format(ctx, report, options)
		if err != nil {
			b.Fatalf("Format failed: %v", err)
		}
	}
}

func BenchmarkReporter_FormatMultiple_10Reports(b *testing.B) {
	ctx := context.Background()
	rep := reporter.NewJSONReporter().(*reporter.JSONReporter)
	reports := make([]*domain.ValidationReport, 10)
	for i := range reports {
		reports[i] = createBenchmarkReport("test.epub", 50)
	}
	options := &ports.ReportOptions{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := rep.FormatMultiple(ctx, reports, options)
		if err != nil {
			b.Fatalf("FormatMultiple failed: %v", err)
		}
	}
}

func BenchmarkReporter_FormatMultiple_100Reports(b *testing.B) {
	ctx := context.Background()
	rep := reporter.NewJSONReporter().(*reporter.JSONReporter)
	reports := make([]*domain.ValidationReport, 100)
	for i := range reports {
		reports[i] = createBenchmarkReport("test.epub", 10)
	}
	options := &ports.ReportOptions{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := rep.FormatMultiple(ctx, reports, options)
		if err != nil {
			b.Fatalf("FormatMultiple failed: %v", err)
		}
	}
}

// Repair Service Benchmarks

func BenchmarkRepairService_EPUB_Preview_SmallReport(b *testing.B) {
	service := epub.NewRepairService()
	ctx := context.Background()
	report := createRepairableReport("test.epub", 5)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := service.Preview(ctx, report)
		if err != nil {
			b.Fatalf("Preview failed: %v", err)
		}
	}
}

func BenchmarkRepairService_EPUB_Preview_MediumReport(b *testing.B) {
	service := epub.NewRepairService()
	ctx := context.Background()
	report := createRepairableReport("test.epub", 50)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := service.Preview(ctx, report)
		if err != nil {
			b.Fatalf("Preview failed: %v", err)
		}
	}
}

func BenchmarkRepairService_EPUB_Preview_LargeReport(b *testing.B) {
	service := epub.NewRepairService()
	ctx := context.Background()
	report := createRepairableReport("test.epub", 200)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := service.Preview(ctx, report)
		if err != nil {
			b.Fatalf("Preview failed: %v", err)
		}
	}
}

func BenchmarkRepairService_EPUB_Apply_Small(b *testing.B) {
	service := epub.NewRepairService()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "epub", "invalid", "missing_title.epub")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	report := createRepairableReport("test.epub", 3)
	preview, err := service.Preview(ctx, report)
	if err != nil {
		b.Fatalf("Preview failed: %v", err)
	}

	tmpDir := b.TempDir()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tmpDir, "output.epub")
		_, err := service.ApplyWithBackup(ctx, testFile, preview, outputPath)
		if err != nil {
			b.Fatalf("Apply failed: %v", err)
		}
		b.StopTimer()
		_ = os.Remove(outputPath)
		b.StartTimer()
	}
}

func BenchmarkRepairService_EPUB_CanRepair(b *testing.B) {
	service := epub.NewRepairService()
	ctx := context.Background()
	err := &domain.ValidationError{
		Code:    "EPUB-MIMETYPE-001",
		Message: "Invalid mimetype",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = service.CanRepair(ctx, err)
	}
}

func BenchmarkRepairService_PDF_Preview_SmallReport(b *testing.B) {
	service := pdf.NewRepairService()
	ctx := context.Background()
	report := createPDFRepairableReport("test.pdf", 5)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := service.Preview(ctx, report)
		if err != nil {
			b.Fatalf("Preview failed: %v", err)
		}
	}
}

func BenchmarkRepairService_PDF_Preview_MediumReport(b *testing.B) {
	service := pdf.NewRepairService()
	ctx := context.Background()
	report := createPDFRepairableReport("test.pdf", 20)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := service.Preview(ctx, report)
		if err != nil {
			b.Fatalf("Preview failed: %v", err)
		}
	}
}

func BenchmarkRepairService_CreateBackup_Small(b *testing.B) {
	service := epub.NewRepairService()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "epub", "valid", "minimal.epub")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	tmpDir := b.TempDir()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		backupPath := filepath.Join(tmpDir, "backup.epub")
		err := service.CreateBackup(ctx, testFile, backupPath)
		if err != nil {
			b.Fatalf("CreateBackup failed: %v", err)
		}
		b.StopTimer()
		_ = os.Remove(backupPath)
		b.StartTimer()
	}
}

func BenchmarkRepairService_CreateBackup_Medium(b *testing.B) {
	service := epub.NewRepairService()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "epub", "valid", "large_100_chapters.epub")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skipf("Test file not found: %s", testFile)
	}

	tmpDir := b.TempDir()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		backupPath := filepath.Join(tmpDir, "backup.epub")
		err := service.CreateBackup(ctx, testFile, backupPath)
		if err != nil {
			b.Fatalf("CreateBackup failed: %v", err)
		}
		b.StopTimer()
		_ = os.Remove(backupPath)
		b.StartTimer()
	}
}

// Helper functions

func createBenchmarkReport(filePath string, errorCount int) *domain.ValidationReport {
	report := &domain.ValidationReport{
		FilePath:       filePath,
		FileType:       "EPUB",
		IsValid:        errorCount == 0,
		Errors:         make([]domain.ValidationError, errorCount),
		Warnings:       make([]domain.ValidationError, errorCount/2),
		Info:           make([]domain.ValidationError, errorCount/4),
		ValidationTime: time.Now(),
		Duration:       100 * time.Millisecond,
		Metadata:       make(map[string]interface{}),
	}

	for i := 0; i < errorCount; i++ {
		report.Errors[i] = domain.ValidationError{
			Code:      "BENCH-001",
			Message:   "Benchmark validation error for testing performance",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
			Location: &domain.ErrorLocation{
				File:   "content.opf",
				Line:   i + 1,
				Column: 5,
				Path:   "OEBPS/content.opf",
			},
			Details: map[string]interface{}{
				"category": "structure",
				"standard": "EPUB3",
				"index":    i,
			},
		}
	}

	for i := 0; i < errorCount/2; i++ {
		report.Warnings[i] = domain.ValidationError{
			Code:      "BENCH-W01",
			Message:   "Benchmark warning",
			Severity:  domain.SeverityWarning,
			Timestamp: time.Now(),
		}
	}

	for i := 0; i < errorCount/4; i++ {
		report.Info[i] = domain.ValidationError{
			Code:      "BENCH-I01",
			Message:   "Benchmark info",
			Severity:  domain.SeverityInfo,
			Timestamp: time.Now(),
		}
	}

	return report
}

func createRepairableReport(filePath string, errorCount int) *domain.ValidationReport {
	report := &domain.ValidationReport{
		FilePath:       filePath,
		FileType:       "EPUB",
		IsValid:        false,
		Errors:         make([]domain.ValidationError, errorCount),
		Warnings:       make([]domain.ValidationError, 0),
		Info:           make([]domain.ValidationError, 0),
		ValidationTime: time.Now(),
		Duration:       100 * time.Millisecond,
		Metadata:       make(map[string]interface{}),
	}

	repairableCodes := []string{
		"EPUB-MIMETYPE-001",
		"EPUB-MIMETYPE-002",
		"EPUB-CONTAINER-001",
		"EPUB-OPF-004",
		"EPUB-OPF-005",
		"EPUB-OPF-006",
		"EPUB-OPF-007",
	}

	for i := 0; i < errorCount; i++ {
		code := repairableCodes[i%len(repairableCodes)]
		report.Errors[i] = domain.ValidationError{
			Code:      code,
			Message:   "Repairable validation error",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
			Location: &domain.ErrorLocation{
				File: "content.opf",
				Path: "OEBPS/content.opf",
			},
			Details: map[string]interface{}{},
		}
	}

	return report
}

func createPDFRepairableReport(filePath string, errorCount int) *domain.ValidationReport {
	report := &domain.ValidationReport{
		FilePath:       filePath,
		FileType:       "PDF",
		IsValid:        false,
		Errors:         make([]domain.ValidationError, errorCount),
		Warnings:       make([]domain.ValidationError, 0),
		Info:           make([]domain.ValidationError, 0),
		ValidationTime: time.Now(),
		Duration:       50 * time.Millisecond,
		Metadata:       make(map[string]interface{}),
	}

	repairableCodes := []string{
		"PDF-TRAILER-001",
		"PDF-TRAILER-003",
	}

	for i := 0; i < errorCount; i++ {
		code := repairableCodes[i%len(repairableCodes)]
		report.Errors[i] = domain.ValidationError{
			Code:      code,
			Message:   "Repairable PDF error",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
			Location: &domain.ErrorLocation{
				File: "document.pdf",
				Path: "document.pdf",
			},
			Details: map[string]interface{}{},
		}
	}

	return report
}
