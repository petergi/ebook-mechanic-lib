package epub

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// Accessibility validation error codes.
const (
	ErrorCodeA11YMissingLang              = "EPUB-A11Y-001"
	ErrorCodeA11YInvalidLang              = "EPUB-A11Y-002"
	ErrorCodeA11YMissingSemanticStructure = "EPUB-A11Y-003"
	ErrorCodeA11YInvalidHeadingHierarchy  = "EPUB-A11Y-004"
	ErrorCodeA11YMissingAltText           = "EPUB-A11Y-005"
	ErrorCodeA11YEmptyAltText             = "EPUB-A11Y-006"
	ErrorCodeA11YInvalidARIARole          = "EPUB-A11Y-007"
	ErrorCodeA11YInvalidARIAAttribute     = "EPUB-A11Y-008"
	ErrorCodeA11YMissingARIALabel         = "EPUB-A11Y-009"
	ErrorCodeA11YInvalidReadingOrder      = "EPUB-A11Y-010"
	ErrorCodeA11YMissingTableHeaders      = "EPUB-A11Y-011"
	ErrorCodeA11YInvalidTableStructure    = "EPUB-A11Y-012"
	ErrorCodeA11YMissingFormLabels        = "EPUB-A11Y-013"
	ErrorCodeA11YInsufficientContrast     = "EPUB-A11Y-014"
	ErrorCodeA11YMediaMissingAlt          = "EPUB-A11Y-015"
	ErrorCodeA11YMediaOverlaySync         = "EPUB-A11Y-016"
	ErrorCodeA11YMissingSkipLinks         = "EPUB-A11Y-017"
	ErrorCodeA11YInvalidLandmarks         = "EPUB-A11Y-018"
	ErrorCodeA11YEmptyHeading             = "EPUB-A11Y-019"
	ErrorCodeA11YSkippedHeadingLevel      = "EPUB-A11Y-020"
)

// WCAG 2.1 and EPUB Accessibility 1.1 constants.
const (
	MinimumScore       = 0
	MaximumScore       = 100
	PassingScore       = 80
	SemanticWeight     = 25
	ARIAWeight         = 20
	AltTextWeight      = 25
	HeadingWeight      = 15
	ReadingOrderWeight = 10
	LangWeight         = 5
)

// Valid ARIA roles (common subset).
var validARIARoles = map[string]bool{
	"alert": true, "alertdialog": true, "application": true, "article": true,
	"banner": true, "button": true, "checkbox": true, "columnheader": true,
	"combobox": true, "complementary": true, "contentinfo": true, "definition": true,
	"dialog": true, "directory": true, "document": true, "feed": true,
	"figure": true, "form": true, "grid": true, "gridcell": true,
	"group": true, "heading": true, "img": true, "link": true,
	"list": true, "listbox": true, "listitem": true, "log": true,
	"main": true, "marquee": true, "math": true, "menu": true,
	"menubar": true, "menuitem": true, "menuitemcheckbox": true, "menuitemradio": true,
	"navigation": true, "none": true, "note": true, "option": true,
	"presentation": true, "progressbar": true, "radio": true, "radiogroup": true,
	"region": true, "row": true, "rowgroup": true, "rowheader": true,
	"scrollbar": true, "search": true, "searchbox": true, "separator": true,
	"slider": true, "spinbutton": true, "status": true, "switch": true,
	"tab": true, "table": true, "tablist": true, "tabpanel": true,
	"term": true, "textbox": true, "timer": true, "toolbar": true,
	"tooltip": true, "tree": true, "treegrid": true, "treeitem": true,
}

// Valid ARIA attributes (common subset).
var validARIAAttributes = map[string]bool{
	"aria-activedescendant": true, "aria-atomic": true, "aria-autocomplete": true,
	"aria-busy": true, "aria-checked": true, "aria-colcount": true,
	"aria-colindex": true, "aria-colspan": true, "aria-controls": true,
	"aria-current": true, "aria-describedby": true, "aria-details": true,
	"aria-disabled": true, "aria-dropeffect": true, "aria-errormessage": true,
	"aria-expanded": true, "aria-flowto": true, "aria-grabbed": true,
	"aria-haspopup": true, "aria-hidden": true, "aria-invalid": true,
	"aria-keyshortcuts": true, "aria-label": true, "aria-labelledby": true,
	"aria-level": true, "aria-live": true, "aria-modal": true,
	"aria-multiline": true, "aria-multiselectable": true, "aria-orientation": true,
	"aria-owns": true, "aria-placeholder": true, "aria-posinset": true,
	"aria-pressed": true, "aria-readonly": true, "aria-relevant": true,
	"aria-required": true, "aria-roledescription": true, "aria-rowcount": true,
	"aria-rowindex": true, "aria-rowspan": true, "aria-selected": true,
	"aria-setsize": true, "aria-sort": true, "aria-valuemax": true,
	"aria-valuemin": true, "aria-valuenow": true, "aria-valuetext": true,
}

// HTML5 semantic elements.
var semanticElements = map[string]bool{
	"article": true, "aside": true, "details": true, "figcaption": true,
	"figure": true, "footer": true, "header": true, "main": true,
	"mark": true, "nav": true, "section": true, "summary": true,
	"time": true,
}

// AccessibilityScore represents the accessibility scoring breakdown.
type AccessibilityScore struct {
	Total               int                    `json:"total"`
	SemanticStructure   int                    `json:"semantic_structure"`
	ARIACompliance      int                    `json:"aria_compliance"`
	AltTextCompleteness int                    `json:"alt_text_completeness"`
	HeadingHierarchy    int                    `json:"heading_hierarchy"`
	ReadingOrder        int                    `json:"reading_order"`
	LanguageDeclaration int                    `json:"language_declaration"`
	Details             map[string]interface{} `json:"details"`
}

// AccessibilityMetadata contains metadata for package document.
type AccessibilityMetadata struct {
	ConformanceClaims     []string               `json:"conformance_claims"`
	AccessModes           []string               `json:"access_modes"`
	AccessModeSufficient  []string               `json:"access_mode_sufficient"`
	AccessibilityFeatures []string               `json:"accessibility_features"`
	AccessibilityHazards  []string               `json:"accessibility_hazards"`
	AccessibilitySummary  string                 `json:"accessibility_summary"`
	CertifiedBy           string                 `json:"certified_by,omitempty"`
	CertifierCredential   string                 `json:"certifier_credential,omitempty"`
	AdditionalMetadata    map[string]interface{} `json:"additional_metadata"`
}

// AccessibilityValidationResult contains accessibility validation details.
type AccessibilityValidationResult struct {
	Valid                  bool
	Errors                 []ValidationError
	Warnings               []ValidationError
	Score                  AccessibilityScore
	Metadata               AccessibilityMetadata
	HasLanguageDeclaration bool
	HasSemanticStructure   bool
	HasARIAAttributes      bool
	ImagesWithAlt          int
	ImagesWithoutAlt       int
	TotalImages            int
	HeadingStructure       []HeadingInfo
	MediaOverlays          []MediaOverlayInfo
	ReadingOrderIssues     int
	ComplianceLevel        string
}

// HeadingInfo represents heading structure information.
type HeadingInfo struct {
	Level   int
	Text    string
	IsEmpty bool
}

// MediaOverlayInfo represents media overlay synchronization information.
type MediaOverlayInfo struct {
	TextElement  string
	AudioElement string
	IsSynced     bool
}

// AccessibilityValidator validates EPUB accessibility compliance.
type AccessibilityValidator struct{}

// NewAccessibilityValidator returns a new accessibility validator.
func NewAccessibilityValidator() *AccessibilityValidator {
	return &AccessibilityValidator{}
}

// ValidateFile validates accessibility from a file path.
func (v *AccessibilityValidator) ValidateFile(filePath string) (*AccessibilityValidationResult, error) {
	file, err := os.Open(filePath) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	return v.Validate(file)
}

// ValidateBytes validates accessibility from in-memory data.
func (v *AccessibilityValidator) ValidateBytes(data []byte) (*AccessibilityValidationResult, error) {
	return v.Validate(strings.NewReader(string(data)))
}

// Validate validates accessibility from an io.Reader.
func (v *AccessibilityValidator) Validate(reader io.Reader) (*AccessibilityValidationResult, error) {
	result := &AccessibilityValidationResult{
		Valid:            true,
		Errors:           make([]ValidationError, 0),
		Warnings:         make([]ValidationError, 0),
		HeadingStructure: make([]HeadingInfo, 0),
		MediaOverlays:    make([]MediaOverlayInfo, 0),
		Score: AccessibilityScore{
			Details: make(map[string]interface{}),
		},
		Metadata: AccessibilityMetadata{
			ConformanceClaims:     make([]string, 0),
			AccessModes:           make([]string, 0),
			AccessModeSufficient:  make([]string, 0),
			AccessibilityFeatures: make([]string, 0),
			AccessibilityHazards:  make([]string, 0),
			AdditionalMetadata:    make(map[string]interface{}),
		},
	}

	doc, parseErr := html.Parse(reader)
	if parseErr != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeContentNotWellFormed,
			Message: "Content document is not well-formed XHTML",
			Details: map[string]interface{}{
				"error": parseErr.Error(),
			},
		})
		return result, nil //nolint:nilerr
	}

	v.validateLanguageDeclaration(doc, result)
	v.validateSemanticStructure(doc, result)
	v.validateARIA(doc, result)
	v.validateImageAltText(doc, result)
	v.validateHeadingHierarchy(doc, result)
	v.validateReadingOrder(doc, result)
	v.validateTables(doc, result)
	v.validateForms(doc, result)
	v.validateMediaElements(doc, result)
	v.validateLandmarks(doc, result)

	v.calculateScore(result)
	v.generateMetadata(result)
	v.determineComplianceLevel(result)

	result.Valid = len(result.Errors) == 0

	return result, nil
}

func (v *AccessibilityValidator) validateLanguageDeclaration(doc *html.Node, result *AccessibilityValidationResult) {
	htmlNode := v.findElement(doc, "html")
	if htmlNode == nil {
		return
	}

	langAttr := v.getAttribute(htmlNode, "lang")
	xmlLangAttr := v.getAttribute(htmlNode, "xml:lang")

	if langAttr == "" && xmlLangAttr == "" {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeA11YMissingLang,
			Message: "HTML element must have lang and/or xml:lang attribute for language declaration",
			Details: map[string]interface{}{},
		})
		return
	}

	result.HasLanguageDeclaration = true

	langValue := langAttr
	if langValue == "" {
		langValue = xmlLangAttr
	}

	if !v.isValidLanguageCode(langValue) {
		result.Warnings = append(result.Warnings, ValidationError{
			Code:    ErrorCodeA11YInvalidLang,
			Message: fmt.Sprintf("Language code '%s' may not be valid", langValue),
			Details: map[string]interface{}{
				"language_code": langValue,
			},
		})
	}
}

func (v *AccessibilityValidator) isValidLanguageCode(lang string) bool {
	if lang == "" {
		return false
	}
	langPattern := regexp.MustCompile(`^[a-z]{2,3}(-[A-Z]{2})?$`)
	return langPattern.MatchString(lang)
}

func (v *AccessibilityValidator) validateSemanticStructure(doc *html.Node, result *AccessibilityValidationResult) {
	semanticCount := 0
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && semanticElements[n.Data] {
			semanticCount++
			result.HasSemanticStructure = true
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	if semanticCount == 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Code:    ErrorCodeA11YMissingSemanticStructure,
			Message: "Document contains no HTML5 semantic elements (article, aside, nav, section, etc.)",
			Details: map[string]interface{}{
				"recommendation": "Use semantic HTML5 elements to improve document structure",
			},
		})
	}

	result.Score.Details["semantic_elements_count"] = semanticCount
}

func (v *AccessibilityValidator) validateARIA(doc *html.Node, result *AccessibilityValidationResult) {
	ariaCount := 0
	invalidRoles := 0
	invalidAttributes := 0
	missingLabels := 0

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			role := v.getAttribute(n, "role")
			if role != "" {
				ariaCount++
				result.HasARIAAttributes = true
				if !validARIARoles[role] {
					invalidRoles++
					result.Errors = append(result.Errors, ValidationError{
						Code:    ErrorCodeA11YInvalidARIARole,
						Message: fmt.Sprintf("Invalid ARIA role '%s' on <%s> element", role, n.Data),
						Details: map[string]interface{}{
							"role":    role,
							"element": n.Data,
						},
					})
				}

				if v.requiresARIALabel(role) && !v.hasARIALabel(n) {
					missingLabels++
					result.Errors = append(result.Errors, ValidationError{
						Code:    ErrorCodeA11YMissingARIALabel,
						Message: fmt.Sprintf("Element with role '%s' requires aria-label or aria-labelledby", role),
						Details: map[string]interface{}{
							"role":    role,
							"element": n.Data,
						},
					})
				}
			}

			for _, attr := range n.Attr {
				if strings.HasPrefix(attr.Key, "aria-") {
					ariaCount++
					result.HasARIAAttributes = true
					if !validARIAAttributes[attr.Key] {
						invalidAttributes++
						result.Warnings = append(result.Warnings, ValidationError{
							Code:    ErrorCodeA11YInvalidARIAAttribute,
							Message: fmt.Sprintf("Unknown ARIA attribute '%s' on <%s> element", attr.Key, n.Data),
							Details: map[string]interface{}{
								"attribute": attr.Key,
								"element":   n.Data,
							},
						})
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	result.Score.Details["aria_attributes_count"] = ariaCount
	result.Score.Details["invalid_aria_roles"] = invalidRoles
	result.Score.Details["invalid_aria_attributes"] = invalidAttributes
	result.Score.Details["missing_aria_labels"] = missingLabels
}

func (v *AccessibilityValidator) requiresARIALabel(role string) bool {
	requiresLabel := map[string]bool{
		"region": true, "form": true, "navigation": true, "search": true,
		"complementary": true, "banner": true, "contentinfo": true,
	}
	return requiresLabel[role]
}

func (v *AccessibilityValidator) hasARIALabel(n *html.Node) bool {
	return v.getAttribute(n, "aria-label") != "" || v.getAttribute(n, "aria-labelledby") != ""
}

func (v *AccessibilityValidator) validateImageAltText(doc *html.Node, result *AccessibilityValidationResult) {
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			result.TotalImages++
			alt := v.getAttribute(n, "alt")
			src := v.getAttribute(n, "src")
			role := v.getAttribute(n, "role")

			switch {
			case role == "presentation" || role == "none":
				result.ImagesWithAlt++
			case alt == "":
				result.ImagesWithoutAlt++
				result.Errors = append(result.Errors, ValidationError{
					Code:    ErrorCodeA11YMissingAltText,
					Message: fmt.Sprintf("Image missing alt attribute: %s", src),
					Details: map[string]interface{}{
						"src": src,
					},
				})
			case strings.TrimSpace(alt) == "":
				result.ImagesWithAlt++
				result.Warnings = append(result.Warnings, ValidationError{
					Code:    ErrorCodeA11YEmptyAltText,
					Message: fmt.Sprintf("Image has empty alt text (may be decorative): %s", src),
					Details: map[string]interface{}{
						"src": src,
					},
				})
			default:
				result.ImagesWithAlt++
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	result.Score.Details["images_with_alt"] = result.ImagesWithAlt
	result.Score.Details["images_without_alt"] = result.ImagesWithoutAlt
	result.Score.Details["total_images"] = result.TotalImages
}

func (v *AccessibilityValidator) validateHeadingHierarchy(doc *html.Node, result *AccessibilityValidationResult) {
	headings := make([]HeadingInfo, 0)
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && len(n.Data) == 2 && n.Data[0] == 'h' && n.Data[1] >= '1' && n.Data[1] <= '6' {
			level := int(n.Data[1] - '0')
			text := v.extractText(n)
			isEmpty := strings.TrimSpace(text) == ""

			headings = append(headings, HeadingInfo{
				Level:   level,
				Text:    text,
				IsEmpty: isEmpty,
			})

			if isEmpty {
				result.Errors = append(result.Errors, ValidationError{
					Code:    ErrorCodeA11YEmptyHeading,
					Message: fmt.Sprintf("Heading <%s> is empty", n.Data),
					Details: map[string]interface{}{
						"level": level,
					},
				})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	result.HeadingStructure = headings

	if len(headings) > 0 {
		if headings[0].Level != 1 {
			result.Warnings = append(result.Warnings, ValidationError{
				Code:    ErrorCodeA11YInvalidHeadingHierarchy,
				Message: fmt.Sprintf("First heading should be <h1>, found <h%d>", headings[0].Level),
				Details: map[string]interface{}{
					"first_heading_level": headings[0].Level,
				},
			})
		}

		for i := 1; i < len(headings); i++ {
			if headings[i].Level > headings[i-1].Level+1 {
				result.Errors = append(result.Errors, ValidationError{
					Code:    ErrorCodeA11YSkippedHeadingLevel,
					Message: fmt.Sprintf("Heading hierarchy skipped from <h%d> to <h%d>", headings[i-1].Level, headings[i].Level),
					Details: map[string]interface{}{
						"from_level": headings[i-1].Level,
						"to_level":   headings[i].Level,
						"text":       headings[i].Text,
					},
				})
			}
		}
	}

	result.Score.Details["heading_count"] = len(headings)
}

func (v *AccessibilityValidator) validateReadingOrder(doc *html.Node, result *AccessibilityValidationResult) {
	bodyNode := v.findElement(doc, "body")
	if bodyNode == nil {
		return
	}

	tabIndexIssues := 0
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			tabIndex := v.getAttribute(n, "tabindex")
			if tabIndex != "" {
				if tabIndexVal := v.parseTabIndex(tabIndex); tabIndexVal > 0 {
					tabIndexIssues++
					result.Warnings = append(result.Warnings, ValidationError{
						Code:    ErrorCodeA11YInvalidReadingOrder,
						Message: fmt.Sprintf("Positive tabindex (%s) disrupts natural reading order on <%s>", tabIndex, n.Data),
						Details: map[string]interface{}{
							"tabindex": tabIndex,
							"element":  n.Data,
						},
					})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(bodyNode)

	result.ReadingOrderIssues = tabIndexIssues
	result.Score.Details["reading_order_issues"] = tabIndexIssues
}

func (v *AccessibilityValidator) parseTabIndex(tabIndex string) int {
	var val int
	_, _ = fmt.Sscanf(tabIndex, "%d", &val)
	return val
}

func (v *AccessibilityValidator) validateTables(doc *html.Node, result *AccessibilityValidationResult) {
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			role := v.getAttribute(n, "role")
			if role == "presentation" || role == "none" {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					traverse(c)
				}
				return
			}

			hasHeaders := false
			hasTh := false
			var checkHeaders func(*html.Node)
			checkHeaders = func(node *html.Node) {
				if node.Type == html.ElementNode {
					if node.Data == "th" {
						hasTh = true
					}
					if node.Data == "td" && v.getAttribute(node, "headers") != "" {
						hasHeaders = true
					}
				}
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					checkHeaders(c)
				}
			}
			checkHeaders(n)

			if !hasTh && !hasHeaders {
				result.Errors = append(result.Errors, ValidationError{
					Code:    ErrorCodeA11YMissingTableHeaders,
					Message: "Data table missing header cells (<th>) or headers attribute",
					Details: map[string]interface{}{},
				})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
}

func (v *AccessibilityValidator) validateForms(doc *html.Node, result *AccessibilityValidationResult) {
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "input" || n.Data == "select" || n.Data == "textarea" {
				inputType := v.getAttribute(n, "type")
				if inputType == "hidden" || inputType == "submit" || inputType == "button" {
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						traverse(c)
					}
					return
				}

				id := v.getAttribute(n, "id")
				ariaLabel := v.getAttribute(n, "aria-label")
				ariaLabelledBy := v.getAttribute(n, "aria-labelledby")
				title := v.getAttribute(n, "title")

				hasLabel := false
				if id != "" {
					hasLabel = v.hasLabelFor(doc, id)
				}

				if !hasLabel && ariaLabel == "" && ariaLabelledBy == "" && title == "" {
					result.Errors = append(result.Errors, ValidationError{
						Code:    ErrorCodeA11YMissingFormLabels,
						Message: fmt.Sprintf("Form control <%s> missing label or aria-label", n.Data),
						Details: map[string]interface{}{
							"element": n.Data,
							"type":    inputType,
						},
					})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
}

func (v *AccessibilityValidator) hasLabelFor(doc *html.Node, forID string) bool {
	found := false
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if found {
			return
		}
		if n.Type == html.ElementNode && n.Data == "label" {
			if v.getAttribute(n, "for") == forID {
				found = true
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	return found
}

func (v *AccessibilityValidator) validateMediaElements(doc *html.Node, result *AccessibilityValidationResult) {
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "audio", "video":
				src := v.getAttribute(n, "src")

				hasCaptions := false
				hasDescription := false
				var checkTracks func(*html.Node)
				checkTracks = func(node *html.Node) {
					if node.Type == html.ElementNode && node.Data == "track" {
						kind := v.getAttribute(node, "kind")
						if kind == "captions" || kind == "subtitles" {
							hasCaptions = true
						}
						if kind == "descriptions" {
							hasDescription = true
						}
					}
					for c := node.FirstChild; c != nil; c = c.NextSibling {
						checkTracks(c)
					}
				}
				checkTracks(n)

				if !hasCaptions && n.Data == "video" {
					result.Warnings = append(result.Warnings, ValidationError{
						Code:    ErrorCodeA11YMediaMissingAlt,
						Message: fmt.Sprintf("Video element missing captions/subtitles: %s", src),
						Details: map[string]interface{}{
							"src": src,
						},
					})
				}

				if !hasDescription && n.Data == "video" {
					result.Warnings = append(result.Warnings, ValidationError{
						Code:    ErrorCodeA11YMediaMissingAlt,
						Message: fmt.Sprintf("Video element missing audio descriptions: %s", src),
						Details: map[string]interface{}{
							"src": src,
						},
					})
				}

			case "object", "embed":
				text := v.extractText(n)
				if strings.TrimSpace(text) == "" && v.getAttribute(n, "aria-label") == "" {
					result.Warnings = append(result.Warnings, ValidationError{
						Code:    ErrorCodeA11YMediaMissingAlt,
						Message: fmt.Sprintf("Embedded media <%s> missing alternative content", n.Data),
						Details: map[string]interface{}{
							"element": n.Data,
						},
					})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
}

func (v *AccessibilityValidator) validateLandmarks(doc *html.Node, result *AccessibilityValidationResult) {
	landmarks := map[string]int{
		"banner":        0,
		"main":          0,
		"navigation":    0,
		"contentinfo":   0,
		"complementary": 0,
	}

	hasMain := false
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "main" {
				hasMain = true
				landmarks["main"]++
			}
			role := v.getAttribute(n, "role")
			if role != "" {
				if _, exists := landmarks[role]; exists {
					landmarks[role]++
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	if !hasMain && landmarks["main"] == 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Code:    ErrorCodeA11YInvalidLandmarks,
			Message: "Document should contain a <main> landmark or role=\"main\"",
			Details: map[string]interface{}{},
		})
	}

	if landmarks["main"] > 1 {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeA11YInvalidLandmarks,
			Message: fmt.Sprintf("Document contains multiple main landmarks (%d)", landmarks["main"]),
			Details: map[string]interface{}{
				"count": landmarks["main"],
			},
		})
	}

	result.Score.Details["landmarks"] = landmarks
}

func (v *AccessibilityValidator) calculateScore(result *AccessibilityValidationResult) {
	langScore := 0
	if result.HasLanguageDeclaration {
		langScore = LangWeight
	}

	semanticScore := 0
	if semanticCount, ok := result.Score.Details["semantic_elements_count"].(int); ok && semanticCount > 0 {
		if semanticCount >= 5 {
			semanticScore = SemanticWeight
		} else {
			semanticScore = (semanticCount * SemanticWeight) / 5
		}
	}

	ariaScore := ARIAWeight
	if invalidRoles, ok := result.Score.Details["invalid_aria_roles"].(int); ok {
		ariaScore -= invalidRoles * 2
	}
	if missingLabels, ok := result.Score.Details["missing_aria_labels"].(int); ok {
		ariaScore -= missingLabels * 3
	}
	if ariaScore < 0 {
		ariaScore = 0
	}

	altTextScore := AltTextWeight
	if result.TotalImages > 0 {
		altTextScore = (result.ImagesWithAlt * AltTextWeight) / result.TotalImages
	}

	headingScore := HeadingWeight
	headingErrors := 0
	for _, err := range result.Errors {
		if err.Code == ErrorCodeA11YEmptyHeading || err.Code == ErrorCodeA11YSkippedHeadingLevel {
			headingErrors++
		}
	}
	headingScore -= headingErrors * 2
	if headingScore < 0 {
		headingScore = 0
	}

	readingOrderScore := ReadingOrderWeight
	if result.ReadingOrderIssues > 0 {
		readingOrderScore -= result.ReadingOrderIssues
		if readingOrderScore < 0 {
			readingOrderScore = 0
		}
	}

	result.Score.LanguageDeclaration = langScore
	result.Score.SemanticStructure = semanticScore
	result.Score.ARIACompliance = ariaScore
	result.Score.AltTextCompleteness = altTextScore
	result.Score.HeadingHierarchy = headingScore
	result.Score.ReadingOrder = readingOrderScore
	result.Score.Total = langScore + semanticScore + ariaScore + altTextScore + headingScore + readingOrderScore

	if result.Score.Total < MinimumScore {
		result.Score.Total = MinimumScore
	}
	if result.Score.Total > MaximumScore {
		result.Score.Total = MaximumScore
	}
}

func (v *AccessibilityValidator) generateMetadata(result *AccessibilityValidationResult) {
	result.Metadata.AccessModes = append(result.Metadata.AccessModes, "textual")

	if result.TotalImages > 0 && result.ImagesWithAlt == result.TotalImages {
		result.Metadata.AccessModes = append(result.Metadata.AccessModes, "visual")
	}

	if result.TotalImages > 0 && result.ImagesWithAlt == result.TotalImages {
		result.Metadata.AccessibilityFeatures = append(result.Metadata.AccessibilityFeatures, "alternativeText")
	}

	if result.HasSemanticStructure {
		result.Metadata.AccessibilityFeatures = append(result.Metadata.AccessibilityFeatures, "structuralNavigation")
	}

	if result.HasARIAAttributes {
		result.Metadata.AccessibilityFeatures = append(result.Metadata.AccessibilityFeatures, "ARIA")
	}

	if len(result.HeadingStructure) > 0 {
		result.Metadata.AccessibilityFeatures = append(result.Metadata.AccessibilityFeatures, "tableOfContents")
	}

	result.Metadata.AccessibilityHazards = append(result.Metadata.AccessibilityHazards, "none")

	result.Metadata.AccessModeSufficient = append(result.Metadata.AccessModeSufficient, "textual")

	if result.Score.Total >= PassingScore {
		result.Metadata.ConformanceClaims = append(result.Metadata.ConformanceClaims, "WCAG 2.1 Level A")
		if result.Score.Total >= 90 {
			result.Metadata.ConformanceClaims = append(result.Metadata.ConformanceClaims, "WCAG 2.1 Level AA")
		}
	}

	summary := fmt.Sprintf("This publication has an accessibility score of %d/100. ", result.Score.Total)
	if len(result.Errors) > 0 {
		summary += fmt.Sprintf("Contains %d accessibility errors. ", len(result.Errors))
	}
	if len(result.Warnings) > 0 {
		summary += fmt.Sprintf("Contains %d accessibility warnings. ", len(result.Warnings))
	}
	if result.TotalImages > 0 {
		summary += fmt.Sprintf("%d of %d images have alternative text. ", result.ImagesWithAlt, result.TotalImages)
	}

	result.Metadata.AccessibilitySummary = strings.TrimSpace(summary)
}

func (v *AccessibilityValidator) determineComplianceLevel(result *AccessibilityValidationResult) {
	switch {
	case result.Score.Total >= 90 && len(result.Errors) == 0:
		result.ComplianceLevel = "WCAG 2.1 AA"
	case result.Score.Total >= PassingScore && len(result.Errors) == 0:
		result.ComplianceLevel = "WCAG 2.1 A"
	case result.Score.Total >= 60:
		result.ComplianceLevel = "Partial"
	default:
		result.ComplianceLevel = "Non-compliant"
	}
}

func (v *AccessibilityValidator) findElement(n *html.Node, tagName string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tagName {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := v.findElement(c, tagName); found != nil {
			return found
		}
	}
	return nil
}

func (v *AccessibilityValidator) getAttribute(n *html.Node, attrName string) string {
	for _, attr := range n.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

func (v *AccessibilityValidator) extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += v.extractText(c)
	}
	return text
}

// ValidateWithContext validates accessibility with reading order context from spine.
func (v *AccessibilityValidator) ValidateWithContext(reader io.Reader, spineOrder []string) (*AccessibilityValidationResult, error) {
	result, err := v.Validate(reader)
	if err != nil {
		return nil, err
	}

	result.Score.Details["spine_order"] = spineOrder

	return result, nil
}
