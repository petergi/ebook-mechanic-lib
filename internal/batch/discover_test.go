package batch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFilesWithDepthAndIgnore(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0750); err != nil {
		t.Fatal(err)
	}

	paths := []string{
		filepath.Join(root, "one.epub"),
		filepath.Join(root, "two.pdf"),
		filepath.Join(nested, "skip.epub"),
		filepath.Join(root, "ignore.pdf"),
	}
	for _, path := range paths {
		if err := os.WriteFile(path, []byte("data"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	files, err := DiscoverFiles([]string{root}, DiscoverOptions{
		MaxDepth:   1,
		Extensions: []string{".epub", ".pdf"},
		Ignore:     []string{"ignore.pdf"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
}

func TestExpandTargets(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "book.epub")
	if err := os.WriteFile(file, []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}

	pattern := filepath.Join(root, "*.epub")
	files, err := ExpandTargets([]string{pattern})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 match, got %d", len(files))
	}
}
