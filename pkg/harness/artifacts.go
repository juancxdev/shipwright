package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ArtifactRoot = ".harness/artifacts"

var artifactDirs = []string{
	"product",
	"project",
	"design",
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
}

var baseDirs = []string{
	".harness",
	".harness/agents",
	".harness/approvals",
	".harness/runs",
	ArtifactRoot,
}

func CreateBaseStructure() error {
	dirs := append([]string{}, baseDirs...)
	for _, dir := range artifactDirs {
		dirs = append(dirs, filepath.Join(ArtifactRoot, filepath.FromSlash(dir)))
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create %s: %w", dir, err)
		}
		gitkeep := filepath.Join(dir, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			if err := os.WriteFile(gitkeep, []byte{}, 0644); err != nil {
				return fmt.Errorf("cannot write %s: %w", gitkeep, err)
			}
		}
	}
	return nil
}

func ArtifactPath(path string) string {
	clean := filepath.ToSlash(filepath.Clean(path))
	if clean == "." || clean == "" {
		return path
	}
	if strings.HasPrefix(clean, ArtifactRoot+"/") || clean == ArtifactRoot {
		return path
	}
	for _, dir := range artifactDirs {
		if clean == dir || strings.HasPrefix(clean, dir+"/") {
			return filepath.Join(ArtifactRoot, filepath.FromSlash(clean))
		}
	}
	return path
}

func ArtifactExists(path string) bool {
	info, err := os.Stat(ArtifactPath(path))
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func CheckArtifacts(paths []string) []string {
	var missing []string
	for _, p := range paths {
		if !ArtifactExists(p) {
			missing = append(missing, p)
		}
	}
	return missing
}

func WriteFile(path, content string) error {
	resolved := ArtifactPath(path)
	dir := filepath.Dir(resolved)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(resolved, []byte(content), 0644)
}

func AppendFile(path, content string) error {
	resolved := ArtifactPath(path)
	dir := filepath.Dir(resolved)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(resolved, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}
