package harness

import (
	"fmt"
	"os"
	"path/filepath"
)

var baseDirs = []string{
	".harness",
	".harness/agents",
	".harness/approvals",
	".harness/runs",
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

func CreateBaseStructure() error {
	for _, dir := range baseDirs {
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

func ArtifactExists(path string) bool {
	info, err := os.Stat(path)
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
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func AppendFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}
