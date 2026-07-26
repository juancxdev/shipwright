package application

import (
	"os"
	"testing"

	projectprofile "shipwright/internal/projectprofile/application"
)

func TestInstallRecommendedSkillPacksWritesLockAndProjectSkills(t *testing.T) {
	chdirTemp(t)
	writeTestFile(t, "package.json", `{"dependencies":{"astro":"latest"}}`)
	writeTestFile(t, "src/pages/index.astro", "---\n---\n<html></html>")
	profile, err := projectprofile.CalibrateProject("astro-site")
	if err != nil {
		t.Fatalf("CalibrateProject: %v", err)
	}
	if err := projectprofile.SaveProjectProfile(profile); err != nil {
		t.Fatalf("SaveProjectProfile: %v", err)
	}
	if _, err := SaveRecommendedSkillPackManifest(profile); err != nil {
		t.Fatalf("SaveRecommendedSkillPackManifest: %v", err)
	}

	result, err := InstallRecommendedSkillPacks()
	if err != nil {
		t.Fatalf("InstallRecommendedSkillPacks: %v", err)
	}
	if len(result.InstalledPacks) == 0 || len(result.InstalledSkills) == 0 {
		t.Fatalf("expected installed packs/skills: %+v", result)
	}
	if _, err := os.Stat(SkillLockJSON); err != nil {
		t.Fatalf("expected lockfile: %v", err)
	}
	if _, err := os.Stat(".harness/skills/frontend-design/SKILL.md"); err != nil {
		t.Fatalf("expected installed frontend-design skill: %v", err)
	}
	registry, err := BuildSkillRegistry()
	if err != nil {
		t.Fatalf("BuildSkillRegistry: %v", err)
	}
	if FindSkill(registry, "frontend-design") == nil {
		t.Fatal("registry should include .harness/skills frontend-design")
	}
}

func TestInstallSkillsFromLocalSourceWritesLock(t *testing.T) {
	chdirTemp(t)
	writeTestFile(t, "external/custom-ui/SKILL.md", `---
name: custom-ui
---
Use custom UI rules.
`)
	result, err := InstallSkillsFromSource("external")
	if err != nil {
		t.Fatalf("InstallSkillsFromSource: %v", err)
	}
	if len(result.InstalledSkills) != 1 || result.InstalledSkills[0] != "custom-ui" {
		t.Fatalf("unexpected installed skills: %+v", result)
	}
	if _, err := os.Stat(".harness/skills/custom-ui/SKILL.md"); err != nil {
		t.Fatalf("expected external skill installed: %v", err)
	}
	lock, err := LoadSkillLock()
	if err != nil {
		t.Fatalf("LoadSkillLock: %v", err)
	}
	if len(lock.Packs) == 0 || len(lock.Skills) == 0 {
		t.Fatalf("expected lock entries: %+v", lock)
	}
}
