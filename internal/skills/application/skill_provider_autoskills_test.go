package application

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportAutoSkillsToOpenCodeCopiesSkills(t *testing.T) {
	chdirTemp(t)
	writeTestFile(t, filepath.Join(".agents", "skills", "external-ui", "SKILL.md"), `---
name: external-ui
description: External UI skill.
---
Use UI well.
`)
	writeTestFile(t, filepath.Join(".agents", "skills", "external-ui", "references", "notes.md"), "notes")
	writeTestFile(t, filepath.Join(".agents", "skills", "not-a-skill", "README.md"), "skip")

	result, err := ImportAutoSkillsToOpenCode()
	if err != nil {
		t.Fatalf("ImportAutoSkillsToOpenCode: %v", err)
	}
	if len(result.Imported) != 1 || result.Imported[0] != "external-ui" {
		t.Fatalf("imported = %+v", result.Imported)
	}
	if len(result.Skipped) != 1 || result.Skipped[0] != "not-a-skill" {
		t.Fatalf("skipped = %+v", result.Skipped)
	}
	if _, err := os.Stat(filepath.Join(".opencode", "skills", "external-ui", "SKILL.md")); err != nil {
		t.Fatalf("expected imported opencode skill: %v", err)
	}
	if _, err := os.Stat(filepath.Join(".opencode", "skills", "external-ui", "references", "notes.md")); err != nil {
		t.Fatalf("expected imported reference file: %v", err)
	}
}
