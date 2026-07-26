package application

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	projectprofile "shipwright/internal/projectprofile/application"
)

type PlannedStackRefreshResult struct {
	ProfileUpdated     bool          `json:"profile_updated"`
	SkillArtifactsDone bool          `json:"skill_artifacts_done"`
	PlannedStacks      []StackSignal `json:"planned_stacks"`
	AssignmentsCount   int           `json:"assignments_count"`
	DigestsCount       int           `json:"digests_count"`
}

type plannedStackRule struct {
	Name    string
	Kind    string
	Aliases []string
}

func RefreshPlannedStackSkillArtifacts() (*PlannedStackRefreshResult, error) {
	profile, changed, err := RefreshProjectProfileFromPlannedArtifacts()
	if err != nil {
		return nil, err
	}

	registry, err := RefreshSkillRegistry()
	if err != nil {
		return nil, err
	}
	assignments, err := BuildSkillAssignments(registry, profile)
	if err != nil {
		return nil, err
	}
	if err := SaveSkillAssignments(assignments); err != nil {
		return nil, err
	}
	digests := BuildSkillDigests(registry, profile)
	if err := SaveSkillDigests(digests); err != nil {
		return nil, err
	}

	return &PlannedStackRefreshResult{
		ProfileUpdated:     changed,
		SkillArtifactsDone: true,
		PlannedStacks:      profile.PlannedStacks,
		AssignmentsCount:   len(assignments.Technologies),
		DigestsCount:       len(digests.Digests),
	}, nil
}

func RefreshProjectProfileFromPlannedArtifacts() (*ProjectProfile, bool, error) {
	profile, err := projectprofile.LoadProjectProfile()
	profileWasMissing := false
	if err != nil {
		profileWasMissing = true
		projectName := "project"
		if cwd, cwdErr := os.Getwd(); cwdErr == nil {
			projectName = filepath.Base(cwd)
		}
		profile, err = projectprofile.CalibrateProject(projectName)
		if err != nil {
			return nil, false, err
		}
	}

	planned := DetectPlannedStacksFromArtifacts()
	if len(planned) == 0 {
		if profileWasMissing {
			if err := projectprofile.SaveProjectProfile(profile); err != nil {
				return nil, false, err
			}
		}
		return profile, profileWasMissing, nil
	}
	merged := uniqueStacks(append(profile.PlannedStacks, planned...))
	changed := profileWasMissing || !stackSignalsEqual(profile.PlannedStacks, merged)
	if changed {
		profile.PlannedStacks = merged
		profile.GeneratedAt = nowISO()
		if err := projectprofile.SaveProjectProfile(profile); err != nil {
			return nil, false, err
		}
	}
	return profile, changed, nil
}

func DetectPlannedStacksFromArtifacts() []StackSignal {
	var signals []StackSignal
	for _, path := range plannedStackArtifactPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		text := string(data)
		for _, rule := range plannedStackRules() {
			if plannedTextMentions(text, rule.Aliases) {
				signals = append(signals, StackSignal{Name: rule.Name, Kind: rule.Kind, Evidence: fmt.Sprintf("planned stack in %s", filepath.ToSlash(path))})
			}
		}
	}
	return uniqueStacks(signals)
}

func plannedStackArtifactPaths() []string {
	return []string{
		".harness/artifacts/architecture/technology-options.md",
		".harness/artifacts/architecture/system-architecture.md",
		".harness/artifacts/project/delivery-plan.md",
		".harness/artifacts/sdd/design.md",
		".harness/artifacts/sdd/tasks.md",
	}
}

func plannedStackRules() []plannedStackRule {
	return []plannedStackRule{
		{Name: "Next.js", Kind: "planned-frontend", Aliases: []string{"next.js", "nextjs", "next js"}},
		{Name: "React", Kind: "planned-frontend", Aliases: []string{"react", "react.js", "reactjs"}},
		{Name: "Vue", Kind: "planned-frontend", Aliases: []string{"vue", "vue.js", "vuejs"}},
		{Name: "Angular", Kind: "planned-frontend", Aliases: []string{"angular"}},
		{Name: "Svelte", Kind: "planned-frontend", Aliases: []string{"svelte", "sveltekit"}},
		{Name: "Vite", Kind: "planned-tooling", Aliases: []string{"vite"}},
		{Name: "Tailwind CSS", Kind: "planned-styling", Aliases: []string{"tailwind", "tailwind css", "tailwindcss"}},
		{Name: "shadcn/ui", Kind: "planned-ui", Aliases: []string{"shadcn", "shadcn/ui", "shadcn ui"}},
		{Name: "Node.js", Kind: "planned-runtime", Aliases: []string{"node.js", "nodejs", "node js"}},
		{Name: "Express", Kind: "planned-backend", Aliases: []string{"express", "express.js", "expressjs"}},
		{Name: "NestJS", Kind: "planned-backend", Aliases: []string{"nestjs", "nest.js", "nest js"}},
		{Name: "Fastify", Kind: "planned-backend", Aliases: []string{"fastify"}},
		{Name: "Go", Kind: "planned-language", Aliases: []string{"go", "golang"}},
		{Name: "Python", Kind: "planned-language", Aliases: []string{"python", "fastapi", "django", "flask"}},
		{Name: "PostgreSQL", Kind: "planned-data", Aliases: []string{"postgresql", "postgres", "postgis"}},
		{Name: "MySQL", Kind: "planned-data", Aliases: []string{"mysql", "mariadb"}},
		{Name: "MongoDB", Kind: "planned-data", Aliases: []string{"mongodb", "mongo db", "mongo"}},
		{Name: "Supabase", Kind: "planned-data", Aliases: []string{"supabase"}},
		{Name: "Prisma", Kind: "planned-data", Aliases: []string{"prisma"}},
		{Name: "Drizzle ORM", Kind: "planned-data", Aliases: []string{"drizzle", "drizzle orm"}},
		{Name: "Zod", Kind: "planned-validation", Aliases: []string{"zod"}},
		{Name: "React Hook Form", Kind: "planned-forms", Aliases: []string{"react hook form", "react-hook-form"}},
		{Name: "Playwright", Kind: "planned-testing", Aliases: []string{"playwright"}},
		{Name: "Vitest", Kind: "planned-testing", Aliases: []string{"vitest"}},
		{Name: "Docker", Kind: "planned-platform", Aliases: []string{"docker", "dockerfile", "docker compose", "docker-compose"}},
	}
}

func plannedTextMentions(text string, aliases []string) bool {
	lower := strings.ToLower(text)
	for _, alias := range aliases {
		if wholePhraseMatch(lower, strings.ToLower(alias)) {
			return true
		}
	}
	return false
}

func wholePhraseMatch(text, phrase string) bool {
	pattern := `(^|[^a-z0-9])` + regexp.QuoteMeta(phrase) + `([^a-z0-9]|$)`
	return regexp.MustCompile(pattern).FindStringIndex(text) != nil
}

func stackSignalsEqual(a, b []StackSignal) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
