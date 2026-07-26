package domain

import (
	"path/filepath"
	"strings"
)

const Root = ".harness/artifacts"

type Layout struct {
	Root         string
	ArtifactDirs []string
	BaseDirs     []string
}

func DefaultLayout() Layout {
	return Layout{
		Root: Root,
		ArtifactDirs: []string{
			"product",
			"project",
			"design",
			"design/baseline",
			"design/source-screenshots",
			"design/stitch",
			"design/stitch/screens",
			"design/stitch/exports",
			"design/stitch/html",
			"design/opendesign",
			"design/openpencil",
			"architecture",
			"contracts",
			"backlog",
			"sdd",
			"knowledge",
			"knowledge/domain",
			"knowledge/architecture",
			"knowledge/decisions",
			"progress",
			"reports",
		},
		BaseDirs: []string{
			".harness",
			".harness/agents",
			".harness/skills",
			".harness/approvals",
			".harness/runs",
			Root,
		},
	}
}

func ResolvePath(layout Layout, path string) string {
	if layout.Root == "" {
		layout = DefaultLayout()
	}
	clean := filepath.ToSlash(filepath.Clean(path))
	if clean == "." || clean == "" {
		return path
	}
	if strings.HasPrefix(clean, layout.Root+"/") || clean == layout.Root {
		return path
	}
	for _, dir := range layout.ArtifactDirs {
		if clean == dir || strings.HasPrefix(clean, dir+"/") {
			return filepath.Join(layout.Root, filepath.FromSlash(clean))
		}
	}
	return path
}

func AllDirs(layout Layout) []string {
	if layout.Root == "" {
		layout = DefaultLayout()
	}
	dirs := append([]string{}, layout.BaseDirs...)
	for _, dir := range layout.ArtifactDirs {
		dirs = append(dirs, filepath.Join(layout.Root, filepath.FromSlash(dir)))
	}
	return dirs
}
