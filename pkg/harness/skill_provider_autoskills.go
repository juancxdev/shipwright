package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const AutoSkillsSourceDir = ".agents/skills"

type AutoSkillsImportResult struct {
	SourceDir string   `json:"source_dir"`
	TargetDir string   `json:"target_dir"`
	Imported  []string `json:"imported"`
	Skipped   []string `json:"skipped,omitempty"`
}

func AutoSkillsAvailable() bool {
	return dirExists(AutoSkillsSourceDir)
}

func ImportAutoSkillsToOpenCode() (*AutoSkillsImportResult, error) {
	result := &AutoSkillsImportResult{SourceDir: AutoSkillsSourceDir, TargetDir: filepath.Join(".opencode", "skills")}
	entries, err := os.ReadDir(AutoSkillsSourceDir)
	if err != nil {
		return result, fmt.Errorf("autoskills source not found at %s", AutoSkillsSourceDir)
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".git" || entry.Name() == "node_modules" {
			continue
		}
		source := filepath.Join(AutoSkillsSourceDir, entry.Name())
		if !ArtifactExists(filepath.Join(source, "SKILL.md")) && !ArtifactExists(filepath.Join(source, "AGENTS.md")) {
			result.Skipped = append(result.Skipped, entry.Name())
			continue
		}
		target := filepath.Join(".opencode", "skills", entry.Name())
		if err := copySkillDir(source, target); err != nil {
			return result, err
		}
		result.Imported = append(result.Imported, entry.Name())
	}
	sort.Strings(result.Imported)
	sort.Strings(result.Skipped)
	return result, nil
}

func copySkillDir(source, target string) error {
	return filepath.WalkDir(source, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(target, 0755)
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(target, rel), 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return WriteFile(filepath.Join(target, rel), string(data))
	})
}
