package epub

import (
	"strings"
	"testing"
)

func TestAccessibilityValidator_ValidateLanguageDeclaration(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		wantError     bool
		wantErrorCode string
		wantValid     bool
	}{
		{
			name: "valid lang attribute",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body><p>Content</p></body>
</html>`,
			wantError: false,
			wantValid: true,
		},
		{
			name: "valid xml:lang attribute",
			html: `<!DOCTYPE html>
<html xml:lang="en">
<head><title>Test</title></head>
<body><p>Content</p></body>
</html>`,
			wantError: false,
			wantValid: true,
		},
		{
			name: "missing language declaration",
			html: `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body><p>Content</p></body>
</html>`,
			wantError:     true,
			wantErrorCode: ErrorCodeA11YMissingLang,
			wantValid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewAccessibilityValidator()
			result, err := validator.ValidateBytes([]byte(tt.html))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.wantValid)
			}

			if tt.wantError {
				found := false
				for _, e := range result.Errors {
					if e.Code == tt.wantErrorCode {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error code %s not found", tt.wantErrorCode)
				}
			}
		})
	}
}

func TestAccessibilityValidator_ValidateImageAltText(t *testing.T) {
	tests := []struct {
		name              string
		html              string
		wantImagesWithAlt int
		wantTotalImages   int
		wantError         bool
	}{
		{
			name: "image with alt text",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body><img src="test.jpg" alt="Test image"/></body>
</html>`,
			wantImagesWithAlt: 1,
			wantTotalImages:   1,
			wantError:         false,
		},
		{
			name: "image without alt text",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body><img src="test.jpg"/></body>
</html>`,
			wantImagesWithAlt: 0,
			wantTotalImages:   1,
			wantError:         true,
		},
		{
			name: "decorative image with role",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body><img src="test.jpg" role="presentation"/></body>
</html>`,
			wantImagesWithAlt: 1,
			wantTotalImages:   1,
			wantError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewAccessibilityValidator()
			result, err := validator.ValidateBytes([]byte(tt.html))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ImagesWithAlt != tt.wantImagesWithAlt {
				t.Errorf("ImagesWithAlt = %d, want %d", result.ImagesWithAlt, tt.wantImagesWithAlt)
			}

			if result.TotalImages != tt.wantTotalImages {
				t.Errorf("TotalImages = %d, want %d", result.TotalImages, tt.wantTotalImages)
			}

			hasAltError := false
			for _, e := range result.Errors {
				if e.Code == ErrorCodeA11YMissingAltText {
					hasAltError = true
					break
				}
			}

			if hasAltError != tt.wantError {
				t.Errorf("has alt text error = %v, want %v", hasAltError, tt.wantError)
			}
		})
	}
}

func TestAccessibilityValidator_ValidateHeadingHierarchy(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		wantError     bool
		wantErrorCode string
	}{
		{
			name: "valid heading hierarchy",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<h1>Title</h1>
<h2>Section</h2>
<h3>Subsection</h3>
</body>
</html>`,
			wantError: false,
		},
		{
			name: "skipped heading level",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<h1>Title</h1>
<h3>Skipped h2</h3>
</body>
</html>`,
			wantError:     true,
			wantErrorCode: ErrorCodeA11YSkippedHeadingLevel,
		},
		{
			name: "empty heading",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<h1></h1>
</body>
</html>`,
			wantError:     true,
			wantErrorCode: ErrorCodeA11YEmptyHeading,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewAccessibilityValidator()
			result, err := validator.ValidateBytes([]byte(tt.html))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantError {
				found := false
				for _, e := range result.Errors {
					if e.Code == tt.wantErrorCode {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error code %s not found", tt.wantErrorCode)
				}
			}
		})
	}
}

func TestAccessibilityValidator_ValidateARIA(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		wantError     bool
		wantErrorCode string
	}{
		{
			name: "valid ARIA role",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<div role="navigation" aria-label="Main navigation">Content</div>
</body>
</html>`,
			wantError: false,
		},
		{
			name: "invalid ARIA role",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<div role="invalid-role">Content</div>
</body>
</html>`,
			wantError:     true,
			wantErrorCode: ErrorCodeA11YInvalidARIARole,
		},
		{
			name: "missing ARIA label on navigation",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<div role="navigation">Content</div>
</body>
</html>`,
			wantError:     true,
			wantErrorCode: ErrorCodeA11YMissingARIALabel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewAccessibilityValidator()
			result, err := validator.ValidateBytes([]byte(tt.html))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantError {
				found := false
				for _, e := range result.Errors {
					if e.Code == tt.wantErrorCode {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error code %s not found", tt.wantErrorCode)
				}
			}
		})
	}
}

func TestAccessibilityValidator_ValidateTables(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		wantError     bool
		wantErrorCode string
	}{
		{
			name: "table with headers",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<table>
<tr><th>Header 1</th><th>Header 2</th></tr>
<tr><td>Data 1</td><td>Data 2</td></tr>
</table>
</body>
</html>`,
			wantError: false,
		},
		{
			name: "table without headers",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<table>
<tr><td>Data 1</td><td>Data 2</td></tr>
</table>
</body>
</html>`,
			wantError:     true,
			wantErrorCode: ErrorCodeA11YMissingTableHeaders,
		},
		{
			name: "presentation table",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<table role="presentation">
<tr><td>Layout 1</td><td>Layout 2</td></tr>
</table>
</body>
</html>`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewAccessibilityValidator()
			result, err := validator.ValidateBytes([]byte(tt.html))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantError {
				found := false
				for _, e := range result.Errors {
					if e.Code == tt.wantErrorCode {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error code %s not found", tt.wantErrorCode)
				}
			} else {
				for _, e := range result.Errors {
					if e.Code == tt.wantErrorCode {
						t.Errorf("unexpected error: %s", e.Message)
					}
				}
			}
		})
	}
}

func TestAccessibilityValidator_ValidateForms(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		wantError     bool
		wantErrorCode string
	}{
		{
			name: "input with label",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<label for="name">Name:</label>
<input type="text" id="name"/>
</body>
</html>`,
			wantError: false,
		},
		{
			name: "input without label",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<input type="text"/>
</body>
</html>`,
			wantError:     true,
			wantErrorCode: ErrorCodeA11YMissingFormLabels,
		},
		{
			name: "input with aria-label",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<input type="text" aria-label="Name"/>
</body>
</html>`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewAccessibilityValidator()
			result, err := validator.ValidateBytes([]byte(tt.html))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantError {
				found := false
				for _, e := range result.Errors {
					if e.Code == tt.wantErrorCode {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error code %s not found", tt.wantErrorCode)
				}
			}
		})
	}
}

func TestAccessibilityValidator_ValidateSemanticStructure(t *testing.T) {
	tests := []struct {
		name              string
		html              string
		wantSemanticCount int
	}{
		{
			name: "document with semantic elements",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<header>Header</header>
<nav>Navigation</nav>
<main>
<article>Article content</article>
<aside>Sidebar</aside>
</main>
<footer>Footer</footer>
</body>
</html>`,
			wantSemanticCount: 6,
		},
		{
			name: "document without semantic elements",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<div>Content</div>
</body>
</html>`,
			wantSemanticCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewAccessibilityValidator()
			result, err := validator.ValidateBytes([]byte(tt.html))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			semanticCount, ok := result.Score.Details["semantic_elements_count"].(int)
			if !ok {
				t.Fatal("semantic_elements_count not found in score details")
			}

			if semanticCount != tt.wantSemanticCount {
				t.Errorf("semantic element count = %d, want %d", semanticCount, tt.wantSemanticCount)
			}
		})
	}
}

func TestAccessibilityValidator_Score(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		wantMinScore  int
		wantMaxScore  int
		wantCompliant bool
	}{
		{
			name: "fully accessible document",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<header><h1>Main Title</h1></header>
<nav role="navigation" aria-label="Main">
<ul><li><a href="#section1">Section 1</a></li></ul>
</nav>
<main role="main">
<article>
<h2>Section 1</h2>
<p>Content with proper structure.</p>
<figure>
<img src="test.jpg" alt="Descriptive text"/>
<figcaption>Caption</figcaption>
</figure>
</article>
</main>
<footer>Footer content</footer>
</body>
</html>`,
			wantMinScore:  80,
			wantMaxScore:  100,
			wantCompliant: true,
		},
		{
			name: "poorly accessible document",
			html: `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<div>
<img src="test.jpg"/>
<h3>Wrong heading level</h3>
</div>
</body>
</html>`,
			wantMinScore:  0,
			wantMaxScore:  50,
			wantCompliant: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewAccessibilityValidator()
			result, err := validator.ValidateBytes([]byte(tt.html))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Score.Total < tt.wantMinScore || result.Score.Total > tt.wantMaxScore {
				t.Errorf("score = %d, want between %d and %d", result.Score.Total, tt.wantMinScore, tt.wantMaxScore)
			}

			if result.Valid != tt.wantCompliant {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.wantCompliant)
			}
		})
	}
}

func TestAccessibilityValidator_Metadata(t *testing.T) {
	html := `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<main>
<h1>Title</h1>
<p>Content</p>
<img src="test.jpg" alt="Test"/>
</main>
</body>
</html>`

	validator := NewAccessibilityValidator()
	result, err := validator.ValidateBytes([]byte(html))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Metadata.AccessModes) == 0 {
		t.Error("expected access modes to be populated")
	}

	if result.Metadata.AccessibilitySummary == "" {
		t.Error("expected accessibility summary to be generated")
	}

	if !strings.Contains(result.Metadata.AccessibilitySummary, "score") {
		t.Error("expected summary to contain score information")
	}
}

func TestAccessibilityValidator_ReadingOrder(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		wantIssue bool
	}{
		{
			name: "no tabindex issues",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<a href="#">Link 1</a>
<a href="#">Link 2</a>
</body>
</html>`,
			wantIssue: false,
		},
		{
			name: "positive tabindex",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<a href="#" tabindex="1">Link 1</a>
<a href="#" tabindex="2">Link 2</a>
</body>
</html>`,
			wantIssue: true,
		},
		{
			name: "negative tabindex is ok",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<div tabindex="-1">Hidden from tab</div>
</body>
</html>`,
			wantIssue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewAccessibilityValidator()
			result, err := validator.ValidateBytes([]byte(tt.html))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			hasReadingOrderIssue := result.ReadingOrderIssues > 0

			if hasReadingOrderIssue != tt.wantIssue {
				t.Errorf("reading order issues = %v, want %v", hasReadingOrderIssue, tt.wantIssue)
			}
		})
	}
}

func TestAccessibilityValidator_ValidateLandmarks(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		wantError     bool
		wantErrorCode string
	}{
		{
			name: "document with main landmark",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<main>Main content</main>
</body>
</html>`,
			wantError: false,
		},
		{
			name: "document without main landmark",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<div>Content</div>
</body>
</html>`,
			wantError: false,
		},
		{
			name: "document with multiple main landmarks",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<main>Main 1</main>
<main>Main 2</main>
</body>
</html>`,
			wantError:     true,
			wantErrorCode: ErrorCodeA11YInvalidLandmarks,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewAccessibilityValidator()
			result, err := validator.ValidateBytes([]byte(tt.html))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantError {
				found := false
				for _, e := range result.Errors {
					if e.Code == tt.wantErrorCode {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error code %s not found", tt.wantErrorCode)
				}
			}
		})
	}
}

func TestAccessibilityValidator_ComplianceLevel(t *testing.T) {
	tests := []struct {
		name                string
		html                string
		wantComplianceLevel string
	}{
		{
			name: "WCAG 2.1 AA compliant",
			html: `<!DOCTYPE html>
<html lang="en">
<head><title>Test</title></head>
<body>
<header><h1>Title</h1></header>
<nav role="navigation" aria-label="Main">Navigation</nav>
<main role="main">
<article>
<h2>Section</h2>
<p>Content</p>
<img src="test.jpg" alt="Description"/>
</article>
</main>
<footer>Footer</footer>
</body>
</html>`,
			wantComplianceLevel: "WCAG 2.1",
		},
		{
			name: "Non-compliant",
			html: `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<img src="test.jpg"/>
</body>
</html>`,
			wantComplianceLevel: "Non-compliant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewAccessibilityValidator()
			result, err := validator.ValidateBytes([]byte(tt.html))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(result.ComplianceLevel, tt.wantComplianceLevel) {
				t.Errorf("compliance level = %s, want to contain %s", result.ComplianceLevel, tt.wantComplianceLevel)
			}
		})
	}
}
