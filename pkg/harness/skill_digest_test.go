package harness

import (
	"os"
	"strings"
	"testing"
)

func TestBuildSkillDigestsMatchesAgentCapabilities(t *testing.T) {
	registry := &SkillRegistry{Skills: []SkillIndex{
		{Name: "frontend-testing", Description: "Frontend testing rules", Path: "skills/frontend-testing/SKILL.md", AppliesTo: []string{"frontend", "testing"}},
		{Name: "go-testing", Description: "Go testing rules", Path: "skills/go-testing/SKILL.md", AppliesTo: []string{"go", "testing"}},
		{Name: "ui-design", Description: "Design responsive UI", Path: "skills/ui-design/SKILL.md", AppliesTo: []string{"design", "frontend"}},
	}}
	profile := &ProjectProfile{
		Languages: []string{"TypeScript"},
		Commands:  ProjectCommands{Test: []DetectedCommand{{Command: "pnpm test", Source: "package.json", Confidence: "high"}}},
		TDD:       TDDCapability{Supported: true, RecommendedMode: "strict"},
	}

	digests := BuildSkillDigests(registry, profile)
	frontend := FindSkillDigest(digests, "frontend-engineer")
	if frontend == nil {
		t.Fatal("frontend digest missing")
	}
	if !digestHasSkill(*frontend, "frontend-testing") {
		t.Fatalf("frontend digest missing frontend-testing: %+v", frontend.RelevantSkills)
	}
	if !digestHasRule(*frontend, "pnpm test") || !digestHasRule(*frontend, "Strict TDD") {
		t.Fatalf("frontend digest missing profile/TDD rules: %+v", frontend.CompactRules)
	}

	designer := FindSkillDigest(digests, "ui-ux-designer")
	if designer == nil || !digestHasSkill(*designer, "ui-design") {
		t.Fatalf("designer digest = %+v", designer)
	}
}

func TestRefreshSkillDigestsWritesArtifacts(t *testing.T) {
	withTempWorkingDir(t)
	if err := CreateBaseStructure(); err != nil {
		t.Fatalf("CreateBaseStructure: %v", err)
	}
	writeSkillRegistryTestFile(t, "skills/frontend/SKILL.md", `---
name: frontend-skill
description: Frontend implementation standards.
---

Trigger: frontend work.
`)
	registry, err := RefreshSkillRegistry()
	if err != nil {
		t.Fatalf("RefreshSkillRegistry: %v", err)
	}
	digests, err := RefreshSkillDigestsFromRegistry(registry)
	if err != nil {
		t.Fatalf("RefreshSkillDigestsFromRegistry: %v", err)
	}
	if FindSkillDigest(digests, "frontend-engineer") == nil {
		t.Fatal("frontend-engineer digest missing")
	}
	markdown, err := os.ReadFile(SkillDigestsMarkdown)
	if err != nil {
		t.Fatalf("read digests markdown: %v", err)
	}
	if !strings.Contains(string(markdown), "Skill Digests") || !strings.Contains(string(markdown), "frontend-engineer") {
		t.Fatalf("markdown missing digest content:\n%s", string(markdown))
	}
}

func digestHasSkill(digest AgentSkillDigest, name string) bool {
	for _, skill := range digest.RelevantSkills {
		if skill.Name == name {
			return true
		}
	}
	return false
}

func digestHasRule(digest AgentSkillDigest, needle string) bool {
	for _, rule := range digest.CompactRules {
		if strings.Contains(rule, needle) {
			return true
		}
	}
	return false
}
