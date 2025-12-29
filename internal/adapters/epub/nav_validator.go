package epub

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"golang.org/x/net/html"
)

// Navigation validation error codes.
const (
	ErrorCodeNavNotWellFormed       = "EPUB-NAV-001"
	ErrorCodeNavMissingTOC          = "EPUB-NAV-002"
	ErrorCodeNavInvalidTOCStructure = "EPUB-NAV-003"
	ErrorCodeNavInvalidLinks        = "EPUB-NAV-004"
	ErrorCodeNavInvalidLandmarks    = "EPUB-NAV-005"
	ErrorCodeNavMissingNavElement   = "EPUB-NAV-006"
)

// Navigation validation constants.
const (
	EPUBNamespace    = "http://www.idpf.org/2007/ops"
	NavTypeTOC       = "toc"
	NavTypeLandmarks = "landmarks"
)

// NavLink represents a single navigation link.
type NavLink struct {
	Href string
	Text string
}

// NavValidationResult contains navigation validation details.
type NavValidationResult struct {
	Valid         bool
	Errors        []ValidationError
	HasTOC        bool
	HasLandmarks  bool
	TOCLinks      []NavLink
	LandmarkLinks []NavLink
}

// NavValidator validates EPUB navigation documents.
type NavValidator struct{}

// NewNavValidator returns a new navigation validator.
func NewNavValidator() *NavValidator {
	return &NavValidator{}
}

// ValidateFile validates a nav document from a file path.
func (v *NavValidator) ValidateFile(filePath string) (*NavValidationResult, error) {
	file, err := os.Open(filePath) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	return v.Validate(file)
}

// ValidateBytes validates a nav document from in-memory data.
func (v *NavValidator) ValidateBytes(data []byte) (*NavValidationResult, error) {
	return v.Validate(strings.NewReader(string(data)))
}

// Validate validates a nav document from an io.Reader.
func (v *NavValidator) Validate(reader io.Reader) (*NavValidationResult, error) {
	result := &NavValidationResult{
		Valid:         true,
		Errors:        make([]ValidationError, 0),
		TOCLinks:      make([]NavLink, 0),
		LandmarkLinks: make([]NavLink, 0),
	}

	doc, parseErr := html.Parse(reader)
	if parseErr != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeNavNotWellFormed,
			Message: "Navigation document is not well-formed XHTML",
			Details: map[string]interface{}{
				"error": parseErr.Error(),
			},
		})
		return result, nil //nolint:nilerr
	}

	navElements := v.findNavElements(doc)

	if len(navElements) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeNavMissingNavElement,
			Message: "Navigation document must contain at least one <nav> element",
			Details: map[string]interface{}{},
		})
		return result, nil
	}

	tocFound := false

	for _, navNode := range navElements {
		epubType := v.getEpubType(navNode)

		switch epubType {
		case NavTypeTOC:
			tocFound = true
			result.HasTOC = true
			v.validateTOCNav(navNode, result)
		case NavTypeLandmarks:
			result.HasLandmarks = true
			v.validateLandmarksNav(navNode, result)
		}
	}

	if !tocFound {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeNavMissingTOC,
			Message: "Navigation document must contain <nav epub:type=\"toc\">",
			Details: map[string]interface{}{},
		})
	}

	return result, nil
}

func (v *NavValidator) findNavElements(n *html.Node) []*html.Node {
	var navElements []*html.Node

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "nav" {
			navElements = append(navElements, node)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)
	return navElements
}

func (v *NavValidator) getEpubType(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "epub:type" || (attr.Namespace == EPUBNamespace && attr.Key == "type") {
			return attr.Val
		}
	}
	return ""
}

func (v *NavValidator) validateTOCNav(navNode *html.Node, result *NavValidationResult) {
	olNode := v.findFirstChild(navNode, "ol")

	if olNode == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeNavInvalidTOCStructure,
			Message: "TOC <nav> element must contain an <ol> element",
			Details: map[string]interface{}{},
		})
		return
	}

	links := v.extractLinks(olNode)
	result.TOCLinks = links

	for _, link := range links {
		if !v.isValidRelativeLink(link.Href) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodeNavInvalidLinks,
				Message: fmt.Sprintf("TOC contains invalid relative link: %s", link.Href),
				Details: map[string]interface{}{
					"href": link.Href,
					"text": link.Text,
				},
			})
		}
	}
}

func (v *NavValidator) validateLandmarksNav(navNode *html.Node, result *NavValidationResult) {
	olNode := v.findFirstChild(navNode, "ol")

	if olNode == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeNavInvalidLandmarks,
			Message: "Landmarks <nav> element must contain an <ol> element",
			Details: map[string]interface{}{},
		})
		return
	}

	links := v.extractLinks(olNode)
	result.LandmarkLinks = links

	for _, link := range links {
		if !v.isValidRelativeLink(link.Href) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodeNavInvalidLinks,
				Message: fmt.Sprintf("Landmarks contains invalid relative link: %s", link.Href),
				Details: map[string]interface{}{
					"href": link.Href,
					"text": link.Text,
				},
			})
		}
	}
}

func (v *NavValidator) findFirstChild(n *html.Node, tagName string) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tagName {
			return c
		}
		if c.Type == html.ElementNode {
			if found := v.findFirstChild(c, tagName); found != nil {
				return found
			}
		}
	}
	return nil
}

func (v *NavValidator) extractLinks(n *html.Node) []NavLink {
	var links []NavLink

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			href := ""
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					href = attr.Val
					break
				}
			}

			text := v.extractText(node)

			links = append(links, NavLink{
				Href: href,
				Text: strings.TrimSpace(text),
			})
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)
	return links
}

func (v *NavValidator) extractText(n *html.Node) string {
	var text string

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.TextNode {
			text += node.Data
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)
	return text
}

func (v *NavValidator) isValidRelativeLink(href string) bool {
	if href == "" {
		return false
	}

	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return false
	}

	if strings.HasPrefix(href, "//") {
		return false
	}

	if strings.HasPrefix(href, "/") {
		return false
	}

	cleaned := path.Clean(href)
	return !strings.HasPrefix(cleaned, "..")
}
