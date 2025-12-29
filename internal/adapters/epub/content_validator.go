package epub

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/net/html"
)

// Content validation error codes.
const (
	ErrorCodeContentNotWellFormed    = "EPUB-CONTENT-001"
	ErrorCodeContentMissingDoctype   = "EPUB-CONTENT-002"
	ErrorCodeContentInvalidDoctype   = "EPUB-CONTENT-003"
	ErrorCodeContentMissingHTML      = "EPUB-CONTENT-004"
	ErrorCodeContentMissingHead      = "EPUB-CONTENT-005"
	ErrorCodeContentMissingBody      = "EPUB-CONTENT-006"
	ErrorCodeContentInvalidNamespace = "EPUB-CONTENT-007"
	ErrorCodeContentInvalidEncoding  = "EPUB-CONTENT-008"
)

// Content validation constants.
const (
	XHTMLNamespace       = "http://www.w3.org/1999/xhtml"
	ExpectedDoctypeHTML5 = "html"
)

// ContentValidationResult contains XHTML validation details.
type ContentValidationResult struct {
	Valid      bool
	Errors     []ValidationError
	HasDoctype bool
	HasHTML    bool
	HasHead    bool
	HasBody    bool
	Namespace  string
}

// ContentValidator validates XHTML content documents.
type ContentValidator struct{}

// NewContentValidator returns a new content validator.
func NewContentValidator() *ContentValidator {
	return &ContentValidator{}
}

// ValidateFile validates content from a file path.
func (v *ContentValidator) ValidateFile(filePath string) (*ContentValidationResult, error) {
	file, err := os.Open(filePath) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	return v.Validate(file)
}

// ValidateBytes validates content from in-memory data.
func (v *ContentValidator) ValidateBytes(data []byte) (*ContentValidationResult, error) {
	return v.Validate(strings.NewReader(string(data)))
}

// Validate validates content from an io.Reader.
func (v *ContentValidator) Validate(reader io.Reader) (*ContentValidationResult, error) {
	result := &ContentValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	tokenizer := html.NewTokenizer(reader)

	foundHTML := false
	foundHead := false
	foundBody := false
	foundDoctype := false
	htmlNamespace := ""

	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if errors.Is(err, io.EOF) {
				break
			}
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodeContentNotWellFormed,
				Message: "Content document is not well-formed XHTML",
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			})
			return result, nil
		}

		switch tokenType {
		case html.DoctypeToken:
			foundDoctype = true
			result.HasDoctype = true
			token := tokenizer.Token()

			if strings.ToLower(strings.TrimSpace(token.Data)) != ExpectedDoctypeHTML5 {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Code:    ErrorCodeContentInvalidDoctype,
					Message: "Content document must have HTML5 DOCTYPE",
					Details: map[string]interface{}{
						"expected": "<!DOCTYPE html>",
						"found":    token.Data,
					},
				})
			}

		case html.StartTagToken:
			token := tokenizer.Token()
			tagName := strings.ToLower(token.Data)

			switch tagName {
			case "html":
				if foundHTML {
					break
				}
				foundHTML = true
				result.HasHTML = true

				for _, attr := range token.Attr {
					if attr.Key == "xmlns" {
						htmlNamespace = attr.Val
						result.Namespace = attr.Val
						break
					}
				}

				if htmlNamespace != XHTMLNamespace {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Code:    ErrorCodeContentInvalidNamespace,
						Message: "HTML element must have correct XHTML namespace",
						Details: map[string]interface{}{
							"expected": XHTMLNamespace,
							"found":    htmlNamespace,
						},
					})
				}
			case "head":
				if !foundHead {
					foundHead = true
					result.HasHead = true
				}
			case "body":
				if !foundBody {
					foundBody = true
					result.HasBody = true
				}
			}
		case html.EndTagToken, html.TextToken, html.SelfClosingTagToken, html.CommentToken, html.ErrorToken:
			continue
		}
	}

	if !foundDoctype {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeContentMissingDoctype,
			Message: "Content document must have a DOCTYPE declaration",
			Details: map[string]interface{}{},
		})
	}

	if !foundHTML {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeContentMissingHTML,
			Message: "Content document must have an <html> element",
			Details: map[string]interface{}{},
		})
	}

	if !foundHead {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeContentMissingHead,
			Message: "Content document must have a <head> element",
			Details: map[string]interface{}{},
		})
	}

	if !foundBody {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeContentMissingBody,
			Message: "Content document must have a <body> element",
			Details: map[string]interface{}{},
		})
	}

	return result, nil
}
