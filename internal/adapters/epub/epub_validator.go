package epub

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
)

const (
	ErrorCodeEPUBRead           = "EPUB-000"
	ErrorCodeEPUBMultipleErrors = "EPUB-999"
)

type EPUBValidatorImpl struct {
	containerValidator *ContainerValidator
	opfValidator       *OPFValidator
	navValidator       *NavValidator
	contentValidator   *ContentValidator
}

func NewEPUBValidator() ports.EPUBValidator {
	return &EPUBValidatorImpl{
		containerValidator: NewContainerValidator(),
		opfValidator:       NewOPFValidator(),
		navValidator:       NewNavValidator(),
		contentValidator:   NewContentValidator(),
	}
}

func (v *EPUBValidatorImpl) ValidateFile(ctx context.Context, filePath string) (*domain.ValidationReport, error) {
	startTime := time.Now()

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat EPUB file: %w", err)
	}

	report, err := v.validateEPUB(ctx, file, fileInfo.Size(), filePath)
	if err != nil {
		return nil, err
	}

	report.Duration = time.Since(startTime)
	return report, nil
}

func (v *EPUBValidatorImpl) ValidateReader(ctx context.Context, reader io.Reader, size int64) (*domain.ValidationReport, error) {
	startTime := time.Now()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read EPUB data: %w", err)
	}

	readerAt := bytes.NewReader(data)
	report, err := v.validateEPUB(ctx, readerAt, int64(len(data)), "")
	if err != nil {
		return nil, err
	}

	report.Duration = time.Since(startTime)
	return report, nil
}

func (v *EPUBValidatorImpl) ValidateStructure(ctx context.Context, filePath string) (*domain.ValidationReport, error) {
	startTime := time.Now()

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat EPUB file: %w", err)
	}

	report := v.createReport(filePath)

	containerResult, err := v.containerValidator.Validate(file, fileInfo.Size())
	if err != nil {
		return nil, fmt.Errorf("container validation failed: %w", err)
	}

	v.aggregateContainerErrors(containerResult, report)

	report.IsValid = len(report.Errors) == 0
	report.Duration = time.Since(startTime)
	return report, nil
}

func (v *EPUBValidatorImpl) ValidateMetadata(ctx context.Context, filePath string) (*domain.ValidationReport, error) {
	startTime := time.Now()

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat EPUB file: %w", err)
	}

	report := v.createReport(filePath)

	zipReader, err := zip.NewReader(file, fileInfo.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to read EPUB as ZIP: %w", err)
	}

	containerResult, err := v.containerValidator.Validate(file, fileInfo.Size())
	if err != nil {
		return nil, fmt.Errorf("container validation failed: %w", err)
	}

	if !containerResult.Valid || len(containerResult.Rootfiles) == 0 {
		v.aggregateContainerErrors(containerResult, report)
		report.IsValid = false
		report.Duration = time.Since(startTime)
		return report, nil
	}

	opfPath := containerResult.Rootfiles[0].FullPath
	opfData, err := v.readFileFromZip(zipReader, opfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read OPF file: %w", err)
	}

	opfResult, err := v.opfValidator.ValidateBytes(opfData)
	if err != nil {
		return nil, fmt.Errorf("OPF validation failed: %w", err)
	}

	v.aggregateOPFErrors(opfResult, opfPath, report)

	report.IsValid = len(report.Errors) == 0
	report.Duration = time.Since(startTime)
	return report, nil
}

func (v *EPUBValidatorImpl) ValidateContent(ctx context.Context, filePath string) (*domain.ValidationReport, error) {
	startTime := time.Now()

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat EPUB file: %w", err)
	}

	report, err := v.validateEPUB(ctx, file, fileInfo.Size(), filePath)
	if err != nil {
		return nil, err
	}

	report.Duration = time.Since(startTime)
	return report, nil
}

func (v *EPUBValidatorImpl) validateEPUB(ctx context.Context, reader io.ReaderAt, size int64, filePath string) (*domain.ValidationReport, error) {
	report := v.createReport(filePath)

	containerResult, err := v.containerValidator.Validate(reader, size)
	if err != nil {
		return nil, fmt.Errorf("container validation failed: %w", err)
	}

	v.aggregateContainerErrors(containerResult, report)

	if !containerResult.Valid || len(containerResult.Rootfiles) == 0 {
		report.IsValid = false
		return report, nil
	}

	zipReader, err := zip.NewReader(reader, size)
	if err != nil {
		return nil, fmt.Errorf("failed to read EPUB as ZIP: %w", err)
	}

	opfPath := containerResult.Rootfiles[0].FullPath
	opfData, err := v.readFileFromZip(zipReader, opfPath)
	if err != nil {
		v.addError(report, ErrorCodeOPFFileNotFound,
			fmt.Sprintf("Failed to read OPF file at %s: %s", opfPath, err.Error()),
			opfPath, nil)
		report.IsValid = false
		return report, nil
	}

	opfResult, err := v.opfValidator.ValidateBytes(opfData)
	if err != nil {
		return nil, fmt.Errorf("OPF validation failed: %w", err)
	}

	v.aggregateOPFErrors(opfResult, opfPath, report)

	if opfResult.Package == nil || len(opfResult.Package.Manifest.Items) == 0 {
		report.IsValid = false
		return report, nil
	}

	opfDir := path.Dir(opfPath)
	v.validateManifestItems(zipReader, opfResult.Package, opfDir, report)

	report.IsValid = len(report.Errors) == 0
	return report, nil
}

func (v *EPUBValidatorImpl) validateManifestItems(zipReader *zip.Reader, pkg *Package, opfDir string, report *domain.ValidationReport) {
	fileMap := make(map[string]*zip.File)
	for _, f := range zipReader.File {
		fileMap[f.Name] = f
	}

	var navPath string
	for _, item := range pkg.Manifest.Items {
		if strings.Contains(item.Properties, "nav") {
			navPath = item.Href
			break
		}
	}

	if navPath != "" {
		fullNavPath := v.resolvePath(opfDir, navPath)
		navData, err := v.readFileFromZip(zipReader, fullNavPath)
		if err != nil {
			v.addError(report, ErrorCodeOPFFileNotFound,
				fmt.Sprintf("Navigation document referenced but not found: %s", fullNavPath),
				fullNavPath, nil)
		} else {
			navResult, err := v.navValidator.ValidateBytes(navData)
			if err != nil {
				v.addError(report, ErrorCodeNavNotWellFormed,
					fmt.Sprintf("Failed to validate navigation document: %s", err.Error()),
					fullNavPath, nil)
			} else {
				v.aggregateNavErrors(navResult, fullNavPath, report)
			}
		}
	}

	spineIDs := make(map[string]bool)
	for _, spineItem := range pkg.Spine.Items {
		spineIDs[spineItem.IDRef] = true
	}

	for _, item := range pkg.Manifest.Items {
		if !v.isContentDocument(item.MediaType) {
			continue
		}

		if !spineIDs[item.ID] && !strings.Contains(item.Properties, "nav") {
			continue
		}

		fullItemPath := v.resolvePath(opfDir, item.Href)
		itemData, err := v.readFileFromZip(zipReader, fullItemPath)
		if err != nil {
			v.addError(report, ErrorCodeOPFFileNotFound,
				fmt.Sprintf("Content document %s (id=%s) not found in EPUB", fullItemPath, item.ID),
				fullItemPath, map[string]interface{}{
					"manifest_id": item.ID,
					"href":        item.Href,
				})
			continue
		}

		if strings.Contains(item.Properties, "nav") {
			continue
		}

		contentResult, err := v.contentValidator.ValidateBytes(itemData)
		if err != nil {
			v.addError(report, ErrorCodeContentNotWellFormed,
				fmt.Sprintf("Failed to validate content document %s: %s", fullItemPath, err.Error()),
				fullItemPath, map[string]interface{}{
					"manifest_id": item.ID,
				})
		} else {
			v.aggregateContentErrors(contentResult, fullItemPath, item.ID, report)
		}
	}
}

func (v *EPUBValidatorImpl) isContentDocument(mediaType string) bool {
	return mediaType == "application/xhtml+xml" ||
		strings.HasPrefix(mediaType, "text/html") ||
		strings.HasPrefix(mediaType, "application/xhtml")
}

func (v *EPUBValidatorImpl) resolvePath(base, relative string) string {
	if base == "" || base == "." {
		return relative
	}
	return path.Join(base, relative)
}

func (v *EPUBValidatorImpl) readFileFromZip(zipReader *zip.Reader, filePath string) ([]byte, error) {
	filePath = strings.TrimPrefix(filePath, "/")

	for _, f := range zipReader.File {
		if f.Name == filePath {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open file: %w", err)
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			return data, nil
		}
	}

	return nil, fmt.Errorf("file not found: %s", filePath)
}

func (v *EPUBValidatorImpl) createReport(filePath string) *domain.ValidationReport {
	return &domain.ValidationReport{
		FilePath:       filePath,
		FileType:       "EPUB",
		IsValid:        true,
		Errors:         make([]domain.ValidationError, 0),
		Warnings:       make([]domain.ValidationError, 0),
		Info:           make([]domain.ValidationError, 0),
		ValidationTime: time.Now(),
		Metadata:       make(map[string]interface{}),
	}
}

func (v *EPUBValidatorImpl) aggregateContainerErrors(result *ValidationResult, report *domain.ValidationReport) {
	for _, err := range result.Errors {
		v.addError(report, err.Code, err.Message, "mimetype / META-INF/container.xml", err.Details)
	}
}

func (v *EPUBValidatorImpl) aggregateOPFErrors(result *OPFValidationResult, opfPath string, report *domain.ValidationReport) {
	for _, err := range result.Errors {
		v.addError(report, err.Code, err.Message, opfPath, err.Details)
	}
}

func (v *EPUBValidatorImpl) aggregateNavErrors(result *NavValidationResult, navPath string, report *domain.ValidationReport) {
	for _, err := range result.Errors {
		v.addError(report, err.Code, err.Message, navPath, err.Details)
	}
}

func (v *EPUBValidatorImpl) aggregateContentErrors(result *ContentValidationResult, contentPath string, manifestID string, report *domain.ValidationReport) {
	for _, err := range result.Errors {
		details := err.Details
		if details == nil {
			details = make(map[string]interface{})
		}
		details["manifest_id"] = manifestID
		v.addError(report, err.Code, err.Message, contentPath, details)
	}
}

func (v *EPUBValidatorImpl) addError(report *domain.ValidationReport, code, message, file string, details map[string]interface{}) {
	filename := filepath.Base(file)

	validationError := domain.ValidationError{
		Code:      code,
		Message:   message,
		Severity:  domain.SeverityError,
		Timestamp: time.Now(),
		Location: &domain.ErrorLocation{
			File: filename,
			Path: file,
		},
		Details: details,
	}

	report.Errors = append(report.Errors, validationError)
}
