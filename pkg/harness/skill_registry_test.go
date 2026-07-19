package harness

import (
	"os"
	"strings"
	"testing"
)

func TestBuildSkillRegistryIndexesProjectSkills(t *testing.T) {
	withTempWorkingDir(t)
	writeSkillRegistryTestFile(t, "skills/testing/SKILL.md", `---
name: testing-pro
summary: ignored
description: Use when writing tests or enforcing TDD.
---

# Testing Pro

Trigger: when writing tests.

Use go test and keep evidence.
`)
	writeSkillRegistryTestFile(t, ".opencode/skills/product-owner/SKILL.md", `---
name: product-owner
description: Product discovery and scope skill.
---

When to use: discovery.
`)

	registry, err := BuildSkillRegistry()
	if err != nil {
		t.Fatalf("BuildSkillRegistry: %v", err)
	}

	if len(registry.Skills) != 2 {
		t.Fatalf("skills = %+v", registry.Skills)
	}
	testingSkill := FindSkill(registry, "testing-pro")
	if testingSkill == nil {
		t.Fatalf("testing-pro not found: %+v", registry.Skills)
	}
	if testingSkill.Source != "project" || !strings.Contains(testingSkill.Description, "TDD") {
		t.Fatalf("testing skill = %+v", testingSkill)
	}
	if !containsSkillTestString(testingSkill.AppliesTo, "testing") {
		t.Fatalf("expected testing applies_to, got %+v", testingSkill.AppliesTo)
	}
}

func TestSaveSkillRegistryWritesJSONAndMarkdown(t *testing.T) {
	withTempWorkingDir(t)
	if err := CreateBaseStructure(); err != nil {
		t.Fatalf("CreateBaseStructure: %v", err)
	}
	writeSkillRegistryTestFile(t, "skills/go/SKILL.md", `---
name: go-testing
description: Go testing patterns.
---

Trigger: when writing Go tests.
`)

	registry, err := RefreshSkillRegistry()
	if err != nil {
		t.Fatalf("RefreshSkillRegistry: %v", err)
	}
	if FindSkill(registry, "go-testing") == nil {
		t.Fatalf("go-testing not indexed: %+v", registry.Skills)
	}

	loaded, err := LoadSkillRegistry()
	if err != nil {
		t.Fatalf("LoadSkillRegistry: %v", err)
	}
	if FindSkill(loaded, "go-testing") == nil {
		t.Fatalf("go-testing not loaded: %+v", loaded.Skills)
	}
	markdown, err := os.ReadFile(SkillRegistryMarkdown)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	if !strings.Contains(string(markdown), "Skill Registry") || !strings.Contains(string(markdown), "go-testing") {
		t.Fatalf("markdown missing registry content:\n%s", string(markdown))
	}
}

func writeSkillRegistryTestFile(t *testing.T, path, content string) {
	t.Helper()
	writeProjectProfileTestFile(t, path, content)
}

func containsSkillTestString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
