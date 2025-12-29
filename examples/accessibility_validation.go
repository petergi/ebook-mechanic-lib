package main

import (
	"fmt"
	"log"

	"github.com/example/project/internal/adapters/epub"
)

func main() {
	demonstrateAccessibilityValidation()
}

func demonstrateAccessibilityValidation() {
	fmt.Println("=== EPUB Accessibility Validation Demo ===\n")

	validator := epub.NewAccessibilityValidator()

	accessibleContent := `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en" xml:lang="en">
<head>
    <title>Accessible Chapter</title>
</head>
<body>
    <header>
        <h1>Chapter 1: Introduction</h1>
    </header>
    <nav role="navigation" aria-label="Chapter Navigation">
        <ul>
            <li><a href="#section1">Section 1</a></li>
            <li><a href="#section2">Section 2</a></li>
        </ul>
    </nav>
    <main role="main">
        <article>
            <h2 id="section1">Getting Started</h2>
            <p>This is an example of accessible content.</p>
            <figure>
                <img src="diagram.jpg" alt="Diagram showing the process flow"/>
                <figcaption>Figure 1: Process Flow</figcaption>
            </figure>
            
            <h2 id="section2">Key Concepts</h2>
            <p>Understanding accessibility is crucial.</p>
            
            <table>
                <tr>
                    <th>Feature</th>
                    <th>Description</th>
                </tr>
                <tr>
                    <td>Alt Text</td>
                    <td>Descriptions for images</td>
                </tr>
                <tr>
                    <td>Semantic HTML</td>
                    <td>Meaningful structure</td>
                </tr>
            </table>
            
            <form>
                <label for="email">Email:</label>
                <input type="email" id="email" aria-required="true"/>
                
                <label for="message">Message:</label>
                <textarea id="message"></textarea>
            </form>
        </article>
    </main>
    <footer>
        <p>Copyright 2024</p>
    </footer>
</body>
</html>`

	inaccessibleContent := `<!DOCTYPE html>
<html>
<head>
    <title>Inaccessible Chapter</title>
</head>
<body>
    <div>
        <h3>Wrong Heading Level</h3>
        <img src="photo.jpg"/>
        <h5>Skipped h4</h5>
        <div role="invalid-role">Content</div>
        <table>
            <tr><td>No Headers</td><td>Bad Table</td></tr>
        </table>
        <input type="text"/>
    </div>
</body>
</html>`

	fmt.Println("--- Validating Accessible Content ---")
	validateAndDisplay(validator, "Accessible Document", accessibleContent)

	fmt.Println("\n--- Validating Inaccessible Content ---")
	validateAndDisplay(validator, "Inaccessible Document", inaccessibleContent)
}

func validateAndDisplay(validator *epub.AccessibilityValidator, name string, content string) {
	result, err := validator.ValidateBytes([]byte(content))
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	fmt.Printf("\n%s Results:\n", name)
	fmt.Println(repeatChar("=", 60))

	fmt.Printf("Overall Score: %d/100\n", result.Score.Total)
	fmt.Printf("Compliance Level: %s\n", result.ComplianceLevel)
	fmt.Printf("Valid: %v\n\n", result.Valid)

	fmt.Println("Score Breakdown:")
	fmt.Printf("  ├─ Language Declaration:  %2d/%2d\n", result.Score.LanguageDeclaration, 5)
	fmt.Printf("  ├─ Semantic Structure:    %2d/%2d\n", result.Score.SemanticStructure, 25)
	fmt.Printf("  ├─ ARIA Compliance:       %2d/%2d\n", result.Score.ARIACompliance, 20)
	fmt.Printf("  ├─ Alt Text Completeness: %2d/%2d\n", result.Score.AltTextCompleteness, 25)
	fmt.Printf("  ├─ Heading Hierarchy:     %2d/%2d\n", result.Score.HeadingHierarchy, 15)
	fmt.Printf("  └─ Reading Order:         %2d/%2d\n\n", result.Score.ReadingOrder, 10)

	if len(result.Errors) > 0 {
		fmt.Printf("Errors (%d):\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Printf("  %d. [%s] %s\n", i+1, err.Code, err.Message)
			if len(err.Details) > 0 {
				fmt.Printf("     Details: %+v\n", err.Details)
			}
		}
		fmt.Println()
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("Warnings (%d):\n", len(result.Warnings))
		for i, warn := range result.Warnings {
			fmt.Printf("  %d. [%s] %s\n", i+1, warn.Code, warn.Message)
		}
		fmt.Println()
	}

	fmt.Println("Statistics:")
	fmt.Printf("  ├─ Images: %d total, %d with alt text, %d without\n",
		result.TotalImages, result.ImagesWithAlt, result.ImagesWithoutAlt)
	fmt.Printf("  ├─ Headings: %d total\n", len(result.HeadingStructure))
	fmt.Printf("  ├─ Semantic Elements: %v\n", result.Score.Details["semantic_elements_count"])
	fmt.Printf("  └─ Reading Order Issues: %d\n\n", result.ReadingOrderIssues)

	if len(result.HeadingStructure) > 0 {
		fmt.Println("Heading Structure:")
		for i, heading := range result.HeadingStructure {
			indent := repeatChar("  ", heading.Level-1)
			text := heading.Text
			if len(text) > 50 {
				text = text[:47] + "..."
			}
			isEmpty := ""
			if heading.IsEmpty {
				isEmpty = " [EMPTY]"
			}
			fmt.Printf("  %d. %s<h%d>%s</h%d>%s\n", i+1, indent, heading.Level, text, heading.Level, isEmpty)
		}
		fmt.Println()
	}

	fmt.Println("Accessibility Metadata:")
	fmt.Printf("  ├─ Conformance Claims: %v\n", result.Metadata.ConformanceClaims)
	fmt.Printf("  ├─ Access Modes: %v\n", result.Metadata.AccessModes)
	fmt.Printf("  ├─ Accessibility Features: %v\n", result.Metadata.AccessibilityFeatures)
	fmt.Printf("  └─ Summary: %s\n", result.Metadata.AccessibilitySummary)

	fmt.Println(repeatChar("=", 60))
}

func repeatChar(char string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += char
	}
	return result
}
