package application

import (
	"os"
	"strings"
	"testing"

	projectprofile "shipwright/internal/projectprofile/application"
)

func TestBuildSkillAssignmentsDetectsFrontendStackAndAgents(t *testing.T) {
	chdirTemp(t)
	writeTestFile(t, "package.json", `{
  "dependencies": {
    "react": "latest",
    "react-dom": "latest",
    "next": "latest",
    "tailwindcss": "latest",
    "@playwright/test": "latest"
  },
  "devDependencies": {
    "typescript": "latest"
  }
}`)
	writeTestFile(t, "tsconfig.json", `{}`)
	writeTestFile(t, "next.config.ts", `export default {}`)
	writeTestFile(t, "playwright.config.ts", `export default {}`)
	writeTestFile(t, "skills/react/SKILL.md", `---
name: react-best-practices
description: React patterns
---
Use React well.
`)

	profile, err := projectprofile.CalibrateProject("frontend-stack")
	if err != nil {
		t.Fatalf("CalibrateProject: %v", err)
	}
	registry, err := BuildSkillRegistry()
	if err != nil {
		t.Fatalf("BuildSkillRegistry: %v", err)
	}
	set, err := BuildSkillAssignments(registry, profile)
	if err != nil {
		t.Fatalf("BuildSkillAssignments: %v", err)
	}
	if !assignmentHasTechnology(set, "nextjs") || !assignmentHasTechnology(set, "react") || !assignmentHasTechnology(set, "playwright") {
		t.Fatalf("expected next/react/playwright technologies: %+v", set.Technologies)
	}
	if !assignmentHasSkillForAgent(set, "react-best-practices", "frontend-engineer") {
		t.Fatalf("expected react-best-practices assigned to frontend-engineer: %+v", set.Skills)
	}
	if !assignmentHasSkillForAgent(set, "playwright-best-practices", "qa-security-reviewer") {
		t.Fatalf("expected playwright skill assigned to QA: %+v", set.Skills)
	}
}

func TestBuildSkillAssignmentsUsesProjectProfileAsSourceOfTruth(t *testing.T) {
	chdirTemp(t)
	writeTestFile(t, "package.json", `{"dependencies":{"react":"latest","next":"latest"}}`)

	set, err := BuildSkillAssignments(&SkillRegistry{}, &ProjectProfile{ProjectName: "empty-profile"})
	if err != nil {
		t.Fatalf("BuildSkillAssignments: %v", err)
	}
	if assignmentHasTechnology(set, "react") || assignmentHasTechnology(set, "nextjs") {
		t.Fatalf("assignments must not scan package.json directly; got technologies: %+v", set.Technologies)
	}
}

func TestRefreshSkillAssignmentsWritesArtifacts(t *testing.T) {
	chdirTemp(t)
	writeTestFile(t, "go.mod", "module example\n")
	writeTestFile(t, "main.go", "package main\n")
	writeTestFile(t, ".harness/skill-registry.json", `{"version":"1","skills":[]}`)

	set, err := RefreshSkillAssignments()
	if err != nil {
		t.Fatalf("RefreshSkillAssignments: %v", err)
	}
	if !assignmentHasTechnology(set, "go") {
		t.Fatalf("expected go assignment: %+v", set.Technologies)
	}
	for _, path := range []string{SkillAssignmentsJSON, SkillAssignmentsMarkdown} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s: %v", path, err)
		}
	}
	markdown, err := os.ReadFile(SkillAssignmentsMarkdown)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	if !strings.Contains(string(markdown), "Go") || !strings.Contains(string(markdown), "go-testing") {
		t.Fatalf("assignment markdown missing go content:\n%s", string(markdown))
	}
}

func assignmentHasTechnology(set *SkillAssignmentSet, id string) bool {
	for _, tech := range set.Technologies {
		if tech.ID == id {
			return true
		}
	}
	return false
}

func assignmentHasSkillForAgent(set *SkillAssignmentSet, skillName, agent string) bool {
	for _, skill := range set.Skills {
		if skill.Name != skillName {
			continue
		}
		for _, got := range skill.Agents {
			if got == agent {
				return true
			}
		}
	}
	return false
}

func TestBuildSkillAssignmentsAddsFrontendQualitySkillsForPlannedUI(t *testing.T) {
	chdirTemp(t)
	writeTestFile(t, "skills/frontend-design/SKILL.md", `---
name: frontend-design
description: Frontend design skill for responsive UI.
---
`)
	writeTestFile(t, "skills/accessibility/SKILL.md", `---
name: accessibility
description: Accessibility skill for frontend and design QA.
---
`)
	writeTestFile(t, "skills/stitch-generate-design/SKILL.md", `---
name: stitch-generate-design
description: Generate design with Google Stitch.
---
`)
	writeTestFile(t, "skills/existing-web-to-openpencil/SKILL.md", `---
name: existing-web-to-openpencil
description: Reverse-engineer web UI into OpenPencil.
---
`)
	writeTestFile(t, "skills/canvas-generate-design/SKILL.md", `---
name: canvas-generate-design
description: Generate canvas views.
---
`)
	writeTestFile(t, "skills/openpencil-generate-design/SKILL.md", `---
name: openpencil-generate-design
description: Generate OpenPencil canvas views.
---
`)
	writeTestFile(t, "skills/design-code-component-map/SKILL.md", `---
name: design-code-component-map
description: Map design components to code components.
---
`)
	registry, err := BuildSkillRegistry()
	if err != nil {
		t.Fatalf("BuildSkillRegistry: %v", err)
	}
	profile := &ProjectProfile{ProjectName: "planned-ui", PlannedStacks: []StackSignal{{Name: "Next.js", Kind: "planned-frontend", Evidence: "test"}}}

	set, err := BuildSkillAssignments(registry, profile)
	if err != nil {
		t.Fatalf("BuildSkillAssignments: %v", err)
	}
	if !assignmentHasTechnology(set, "frontend-ui-quality") {
		t.Fatalf("expected frontend-ui-quality assignment: %+v", set.Technologies)
	}
	if !assignmentHasSkillForAgent(set, "frontend-design", "ui-ux-designer") {
		t.Fatalf("expected frontend-design assigned to ui-ux-designer: %+v", set.Skills)
	}
	if !assignmentHasSkillForAgent(set, "stitch-generate-design", "ui-ux-designer") {
		t.Fatalf("expected stitch-generate-design assigned to ui-ux-designer: %+v", set.Skills)
	}
	if !assignmentHasSkillForAgent(set, "existing-web-to-openpencil", "ui-ux-designer") {
		t.Fatalf("expected existing-web-to-openpencil assigned to ui-ux-designer: %+v", set.Skills)
	}
	if !assignmentHasSkillForAgent(set, "canvas-generate-design", "ui-ux-designer") {
		t.Fatalf("expected canvas-generate-design assigned to ui-ux-designer: %+v", set.Skills)
	}
	if !assignmentHasSkillForAgent(set, "openpencil-generate-design", "ui-ux-designer") {
		t.Fatalf("expected openpencil-generate-design assigned to ui-ux-designer: %+v", set.Skills)
	}
	if !assignmentHasSkillForAgent(set, "design-code-component-map", "ui-ux-designer") {
		t.Fatalf("expected design-code-component-map assigned to ui-ux-designer: %+v", set.Skills)
	}
	if !assignmentHasSkillForAgent(set, "design-code-component-map", "frontend-engineer") {
		t.Fatalf("expected design-code-component-map assigned to frontend-engineer: %+v", set.Skills)
	}
	if !assignmentHasSkillForAgent(set, "accessibility", "qa-security-reviewer") {
		t.Fatalf("expected accessibility assigned to QA: %+v", set.Skills)
	}
}
