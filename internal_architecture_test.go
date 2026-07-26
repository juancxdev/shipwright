package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInternalPackagesDoNotImportForbiddenOuterLayers(t *testing.T) {
	root := "internal"
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if strings.HasPrefix(filepath.ToSlash(path), "internal/app/harness/") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		forbidden := []string{
			"shipwright/pkg/",
			"shipwright/internal/app/harness",
		}
		for _, imp := range forbidden {
			if strings.Contains(content, imp) {
				t.Fatalf("internal package imports forbidden outer layer %q: %s", imp, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal packages: %v", err)
	}
}
