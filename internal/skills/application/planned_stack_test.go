package application

import (
	"os"
	"strings"
	"testing"

	projectprofile "shipwright/internal/projectprofile/application"
)

func TestRefreshPlannedStackSkillArtifactsUsesTechnicalLeadArtifacts(t *testing.T) {
	chdirTemp(t)
	profile := &ProjectProfile{Version: projectprofile.ProjectProfileVersion, GeneratedAt: nowISO(), ProjectName: "greenfield", Root: "."}
	if err := projectprofile.SaveProjectProfile(profile); err != nil {
		t.Fatalf("SaveProjectProfile: %v", err)
	}
	writeTestFile(t, ".harness/artifacts/architecture/system-architecture.md", `# System Architecture

## Technology stack

- **Frontend**: Next.js + React + Tailwind CSS
- **Backend**: Go API
- **Database**: PostgreSQL
- **Testing**: Playwright
- **Deployment**: Docker
`)
	writeTestFile(t, "skills/next/SKILL.md", `---
name: next-best-practices
description: Next.js best practices
---
Use Next.js well.
`)
	writeTestFile(t, "skills/go/SKILL.md", `---
name: go-testing
description: Go testing patterns
---
Use go test.
`)

	result, err := RefreshPlannedStackSkillArtifacts()
	if err != nil {
		t.Fatalf("RefreshPlannedStackSkillArtifacts: %v", err)
	}
	if !result.ProfileUpdated {
		t.Fatal("expected profile to be updated from planned stack artifacts")
	}

	loaded, err := projectprofile.LoadProjectProfile()
	if err != nil {
		t.Fatalf("LoadProjectProfile: %v", err)
	}
	assertStack(t, loaded.PlannedStacks, "Next.js")
	assertStack(t, loaded.PlannedStacks, "Go")
	assertStack(t, loaded.PlannedStacks, "PostgreSQL")
	assertStack(t, loaded.PlannedStacks, "Playwright")

	assignments, err := LoadSkillAssignments()
	if err != nil {
		t.Fatalf("LoadSkillAssignments: %v", err)
	}
	for _, tech := range []string{"nextjs", "react", "tailwind", "go", "postgresql", "playwright", "docker"} {
		if !assignmentHasTechnology(assignments, tech) {
			t.Fatalf("expected planned tech %s in assignments: %+v", tech, assignments.Technologies)
		}
	}
	if !assignmentHasSkillForAgent(assignments, "next-best-practices", "frontend-engineer") {
		t.Fatalf("expected next skill assigned to frontend: %+v", assignments.Skills)
	}
	if !assignmentHasSkillForAgent(assignments, "go-testing", "backend-engineer") {
		t.Fatalf("expected go-testing assigned to backend: %+v", assignments.Skills)
	}

	markdown, err := os.ReadFile(projectprofile.ProjectProfileMarkdown)
	if err != nil {
		t.Fatalf("read profile markdown: %v", err)
	}
	if !strings.Contains(string(markdown), "## Planned stack") || !strings.Contains(string(markdown), "Next.js") {
		t.Fatalf("profile markdown missing planned stack:\n%s", string(markdown))
	}
}

func TestBuildSkillAssignmentsUsesPlannedStacks(t *testing.T) {
	chdirTemp(t)
	profile := &ProjectProfile{ProjectName: "planned", PlannedStacks: []StackSignal{{Name: "NestJS", Kind: "planned-backend", Evidence: "test"}, {Name: "MongoDB", Kind: "planned-data", Evidence: "test"}}}

	set, err := BuildSkillAssignments(&SkillRegistry{}, profile)
	if err != nil {
		t.Fatalf("BuildSkillAssignments: %v", err)
	}
	if !assignmentHasTechnology(set, "nestjs") || !assignmentHasTechnology(set, "mongodb") {
		t.Fatalf("expected planned stack assignments: %+v", set.Technologies)
	}
}
