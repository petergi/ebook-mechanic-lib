package batch

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type DiscoverOptions struct {
	MaxDepth   int
	Extensions []string
	Ignore     []string
}

func ExpandTargets(targets []string) ([]string, error) {
	var expanded []string
	for _, target := range targets {
		if hasGlob(target) {
			matches, err := filepath.Glob(target)
			if err != nil {
				return nil, err
			}
			expanded = append(expanded, matches...)
			continue
		}
		expanded = append(expanded, target)
	}
	return expanded, nil
}

func DiscoverFiles(targets []string, opts DiscoverOptions) ([]string, error) {
	seen := make(map[string]struct{})
	var files []string

	normalizedExts := normalizeExts(opts.Extensions)
	ignore := opts.Ignore

	for _, target := range targets {
		info, err := os.Stat(target)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			paths, err := walkDir(target, opts.MaxDepth, normalizedExts, ignore)
			if err != nil {
				return nil, err
			}
			for _, path := range paths {
				if _, ok := seen[path]; ok {
					continue
				}
				seen[path] = struct{}{}
				files = append(files, path)
			}
			continue
		}

		if !matchesExt(target, normalizedExts) || matchesIgnore(target, ignore) {
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		files = append(files, target)
	}

	return files, nil
}

func walkDir(root string, maxDepth int, exts, ignore []string) ([]string, error) {
	var files []string
	rootDepth := strings.Count(filepath.Clean(root), string(filepath.Separator))

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if maxDepth >= 0 {
				depth := strings.Count(filepath.Clean(path), string(filepath.Separator)) - rootDepth
				if depth > maxDepth {
					return filepath.SkipDir
				}
			}
			if matchesIgnore(path, ignore) {
				return filepath.SkipDir
			}
			return nil
		}

		if matchesIgnore(path, ignore) {
			return nil
		}
		if !matchesExt(path, exts) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}

	return files, nil
}

func normalizeExts(exts []string) []string {
	if len(exts) == 0 {
		return nil
	}
	out := make([]string, 0, len(exts))
	for _, ext := range exts {
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		out = append(out, strings.ToLower(ext))
	}
	return out
}

func matchesExt(path string, exts []string) bool {
	if len(exts) == 0 {
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	for _, allowed := range exts {
		if ext == allowed {
			return true
		}
	}
	return false
}

func matchesIgnore(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}
	base := filepath.Base(path)
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		if ok, _ := filepath.Match(pattern, path); ok {
			return true
		}
		if ok, _ := filepath.Match(pattern, base); ok {
			return true
		}
	}
	return false
}

func hasGlob(path string) bool {
	return strings.ContainsAny(path, "*?[")
}
