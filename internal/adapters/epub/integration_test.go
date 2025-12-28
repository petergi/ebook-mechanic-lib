package epub_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/project/internal/adapters/epub"
)

func TestContainerValidator_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	validator := epub.NewContainerValidator()

	testCases := []struct {
		name        string
		fixtureFile string
		expectValid bool
	}{
		{
			name:        "Valid EPUB fixture",
			fixtureFile: "valid.epub",
			expectValid: true,
		},
		{
			name:        "Invalid EPUB fixture",
			fixtureFile: "invalid.epub",
			expectValid: false,
		},
	}

	fixtureDir := filepath.Join("..", "..", "..", "testdata", "epub")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixturePath := filepath.Join(fixtureDir, tc.fixtureFile)

			if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
				t.Skipf("Fixture file %s does not exist, run 'go run %s/generate_fixtures.go %s' to create it",
					tc.fixtureFile, fixtureDir, fixtureDir)
				return
			}

			result, err := validator.ValidateFile(fixturePath)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.Valid != tc.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v", tc.expectValid, result.Valid)
				if !result.Valid {
					t.Logf("Validation errors:")
					for _, e := range result.Errors {
						t.Logf("  - [%s] %s", e.Code, e.Message)
					}
				}
			}
		})
	}
}
