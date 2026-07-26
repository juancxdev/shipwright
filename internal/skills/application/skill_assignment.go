package application

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const SkillAssignmentsJSON = ".harness/skill-assignments.json"
const SkillAssignmentsMarkdown = ".harness/skill-assignments.md"
const SkillAssignmentsVersion = "1"

type SkillAssignmentSet struct {
	Version      string                 `json:"version"`
	GeneratedAt  string                 `json:"generated_at"`
	ProjectRoot  string                 `json:"project_root"`
	Technologies []AssignedTechnology   `json:"technologies"`
	Combos       []AssignedTechnology   `json:"combos,omitempty"`
	Skills       []AssignedSkill        `json:"skills"`
	Agents       []AgentSkillAssignment `json:"agents"`
	Warnings     []string               `json:"warnings,omitempty"`
}

type AssignedTechnology struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Kind     string   `json:"kind"`
	Evidence []string `json:"evidence"`
	Skills   []string `json:"skills"`
}

type AssignedSkill struct {
	Name         string   `json:"name"`
	Recommended  bool     `json:"recommended"`
	Installed    bool     `json:"installed"`
	LocalPath    string   `json:"local_path,omitempty"`
	Technologies []string `json:"technologies"`
	Agents       []string `json:"agents"`
	Reason       string   `json:"reason"`
}

type AgentSkillAssignment struct {
	Agent  string   `json:"agent"`
	Skills []string `json:"skills"`
	Rules  []string `json:"rules"`
}

type skillTechRule struct {
	ID          string
	Name        string
	Kind        string
	Packages    []string
	PackageLike []string
	Files       []string
	Extensions  []string
	StackNames  []string
	Skills      []string
}

type skillComboRule struct {
	ID       string
	Name     string
	Requires []string
	Skills   []string
}

func RefreshSkillAssignments() (*SkillAssignmentSet, error) {
	registry, _ := LoadSkillRegistry()
	profile, err := loadOrCalibrateProjectProfileForAssignments()
	if err != nil {
		return nil, err
	}
	set, err := BuildSkillAssignments(registry, profile)
	if err != nil {
		return nil, err
	}
	if err := SaveSkillAssignments(set); err != nil {
		return nil, err
	}
	return set, nil
}

func RefreshSkillAssignmentsFromRegistry(registry *SkillRegistry) (*SkillAssignmentSet, error) {
	profile, err := loadOrCalibrateProjectProfileForAssignments()
	if err != nil {
		return nil, err
	}
	set, err := BuildSkillAssignments(registry, profile)
	if err != nil {
		return nil, err
	}
	if err := SaveSkillAssignments(set); err != nil {
		return nil, err
	}
	return set, nil
}

func BuildSkillAssignments(registry *SkillRegistry, profile *ProjectProfile) (*SkillAssignmentSet, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	ctx := buildSkillDetectContext(profile)
	set := &SkillAssignmentSet{Version: SkillAssignmentsVersion, GeneratedAt: nowISO(), ProjectRoot: root}

	matchedIDs := map[string]bool{}
	for _, rule := range skillTechnologyRules() {
		if evidence := matchSkillTechnology(rule, ctx); len(evidence) > 0 {
			matchedIDs[rule.ID] = true
			set.Technologies = append(set.Technologies, AssignedTechnology{ID: rule.ID, Name: rule.Name, Kind: rule.Kind, Evidence: sortedUnique(evidence), Skills: sortedUnique(rule.Skills)})
		}
	}

	for _, combo := range skillComboRules() {
		ok := true
		for _, required := range combo.Requires {
			if !matchedIDs[required] {
				ok = false
				break
			}
		}
		if ok {
			set.Combos = append(set.Combos, AssignedTechnology{ID: combo.ID, Name: combo.Name, Kind: "combo", Evidence: combo.Requires, Skills: sortedUnique(combo.Skills)})
		}
	}

	set.Technologies = append(set.Technologies, MatchingSkillPacks(profile)...)

	set.Skills = buildAssignedSkills(set.Technologies, set.Combos, registry)
	set.Agents = buildAgentAssignments(set.Skills)
	if len(set.Technologies) == 0 {
		set.Warnings = append(set.Warnings, "no stack-specific skill assignments detected; use baseline Shipwright agent skills")
	}
	return set, nil
}

func SaveSkillAssignments(set *SkillAssignmentSet) error {
	if set == nil {
		return fmt.Errorf("skill assignments are nil")
	}
	data, err := json.MarshalIndent(set, "", "  ")
	if err != nil {
		return err
	}
	if err := writeFile(SkillAssignmentsJSON, string(data)+"\n"); err != nil {
		return err
	}
	return writeFile(SkillAssignmentsMarkdown, RenderSkillAssignmentsMarkdown(set))
}

func LoadSkillAssignments() (*SkillAssignmentSet, error) {
	data, err := os.ReadFile(SkillAssignmentsJSON)
	if err != nil {
		return nil, err
	}
	var set SkillAssignmentSet
	if err := json.Unmarshal(data, &set); err != nil {
		return nil, err
	}
	return &set, nil
}

func RenderSkillAssignmentsMarkdown(set *SkillAssignmentSet) string {
	var sb strings.Builder
	sb.WriteString("# Skill Assignments\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n", set.GeneratedAt))
	sb.WriteString(fmt.Sprintf("**Project root:** `%s`\n", set.ProjectRoot))
	sb.WriteString(fmt.Sprintf("**Technologies:** %d\n", len(set.Technologies)))
	sb.WriteString(fmt.Sprintf("**Recommended skills:** %d\n\n", len(set.Skills)))

	sb.WriteString("## Detected technologies\n\n")
	if len(set.Technologies) == 0 {
		sb.WriteString("No stack-specific technologies detected.\n\n")
	} else {
		for _, tech := range set.Technologies {
			sb.WriteString(fmt.Sprintf("### %s\n\n", tech.Name))
			sb.WriteString(fmt.Sprintf("- ID: `%s`\n", tech.ID))
			sb.WriteString(fmt.Sprintf("- Kind: `%s`\n", tech.Kind))
			if len(tech.Evidence) > 0 {
				sb.WriteString(fmt.Sprintf("- Evidence: `%s`\n", strings.Join(tech.Evidence, "`, `")))
			}
			if len(tech.Skills) > 0 {
				sb.WriteString("- Skills:\n")
				for _, skill := range tech.Skills {
					sb.WriteString(fmt.Sprintf("  - `%s`\n", skill))
				}
			}
			sb.WriteString("\n")
		}
	}

	if len(set.Combos) > 0 {
		sb.WriteString("## Detected combos\n\n")
		for _, combo := range set.Combos {
			sb.WriteString(fmt.Sprintf("- **%s** (`%s`) from `%s` → skills: `%s`\n", combo.Name, combo.ID, strings.Join(combo.Evidence, ", "), strings.Join(combo.Skills, "`, `")))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Recommended skills\n\n")
	if len(set.Skills) == 0 {
		sb.WriteString("No stack-specific recommendations.\n\n")
	} else {
		for _, skill := range set.Skills {
			status := "missing"
			if skill.Installed {
				status = "installed"
			}
			sb.WriteString(fmt.Sprintf("- `%s` — %s; agents: `%s`; technologies: `%s`\n", skill.Name, status, strings.Join(skill.Agents, ", "), strings.Join(skill.Technologies, ", ")))
			if skill.LocalPath != "" {
				sb.WriteString(fmt.Sprintf("  - local: `%s`\n", skill.LocalPath))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Agent assignment\n\n")
	for _, agent := range set.Agents {
		sb.WriteString(fmt.Sprintf("### %s\n\n", agent.Agent))
		if len(agent.Skills) == 0 {
			sb.WriteString("No stack-specific assigned skills.\n\n")
			continue
		}
		for _, skill := range agent.Skills {
			sb.WriteString(fmt.Sprintf("- `%s`\n", skill))
		}
		if len(agent.Rules) > 0 {
			sb.WriteString("\nRules:\n")
			for _, rule := range agent.Rules {
				sb.WriteString(fmt.Sprintf("- %s\n", rule))
			}
		}
		sb.WriteString("\n")
	}

	if len(set.Warnings) > 0 {
		sb.WriteString("## Warnings\n\n")
		for _, warning := range set.Warnings {
			sb.WriteString(fmt.Sprintf("- %s\n", warning))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## How agents must use this\n\n")
	sb.WriteString("- Use this file to decide which stack-specific skills should guide each role.\n")
	sb.WriteString("- Installed skills can be loaded from `.harness/skill-registry.md`; missing recommendations should be treated as gaps, not silently assumed.\n")
	sb.WriteString("- Re-run `shipwright skills refresh` after changing dependencies, project structure, or installed skills.\n")
	return sb.String()
}

type skillDetectContext struct {
	Files      map[string]bool
	Stacks     map[string]bool
	Languages  map[string]bool
	Repository RepositoryProfile
	Structure  ProjectStructure
}

func loadOrCalibrateProjectProfileForAssignments() (*ProjectProfile, error) {
	profile, _, err := RefreshProjectProfileFromPlannedArtifacts()
	return profile, err
}

func buildSkillDetectContext(profile *ProjectProfile) skillDetectContext {
	ctx := skillDetectContext{Files: map[string]bool{}, Stacks: map[string]bool{}, Languages: map[string]bool{}}
	if profile == nil {
		return ctx
	}
	ctx.Repository = profile.Repository
	ctx.Structure = profile.Structure
	for _, stack := range profile.Stacks {
		ctx.Stacks[strings.ToLower(stack.Name)] = true
	}
	for _, stack := range profile.PlannedStacks {
		ctx.Stacks[strings.ToLower(stack.Name)] = true
	}
	for _, lang := range profile.Languages {
		ctx.Languages[strings.ToLower(lang)] = true
	}
	for _, file := range profile.FilesScanned {
		ctx.Files[strings.ToLower(filepath.ToSlash(file))] = true
	}
	for _, file := range profile.ExistingArtifacts {
		ctx.Files[strings.ToLower(filepath.ToSlash(file))] = true
	}
	return ctx
}

func matchSkillTechnology(rule skillTechRule, ctx skillDetectContext) []string {
	var evidence []string
	for _, file := range rule.Files {
		if ctx.Files[strings.ToLower(filepath.ToSlash(file))] {
			evidence = append(evidence, "profile-file:"+file)
		}
	}
	for _, stack := range rule.StackNames {
		if ctx.Stacks[strings.ToLower(stack)] || ctx.Languages[strings.ToLower(stack)] {
			evidence = append(evidence, "profile-stack:"+stack)
		}
	}
	if rule.ID == "docker" {
		if ctx.Repository.Docker {
			evidence = append(evidence, "profile-repository:Docker")
		}
		if ctx.Repository.DockerCompose {
			evidence = append(evidence, "profile-repository:Docker Compose")
		}
	}
	return sortedUnique(evidence)
}

func buildAssignedSkills(techs []AssignedTechnology, combos []AssignedTechnology, registry *SkillRegistry) []AssignedSkill {
	bySkill := map[string]*AssignedSkill{}
	add := func(skill string, tech AssignedTechnology) {
		normalized := strings.TrimSpace(skill)
		if normalized == "" {
			return
		}
		item := bySkill[normalized]
		if item == nil {
			item = &AssignedSkill{Name: normalized, Recommended: true}
			bySkill[normalized] = item
		}
		item.Technologies = append(item.Technologies, tech.ID)
		item.Agents = append(item.Agents, agentsForRecommendedSkill(normalized)...)
	}
	for _, tech := range techs {
		for _, skill := range tech.Skills {
			add(skill, tech)
		}
	}
	for _, combo := range combos {
		for _, skill := range combo.Skills {
			add(skill, combo)
		}
	}
	for _, item := range bySkill {
		item.Technologies = sortedUnique(item.Technologies)
		item.Agents = sortedUnique(item.Agents)
		item.Installed, item.LocalPath = recommendedSkillInstalled(item.Name, registry)
		if item.Installed {
			item.Reason = "recommended and available locally"
		} else {
			item.Reason = "recommended by detected stack but not found in local skill registry"
		}
	}
	items := make([]AssignedSkill, 0, len(bySkill))
	for _, item := range bySkill {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items
}

func buildAgentAssignments(skills []AssignedSkill) []AgentSkillAssignment {
	byAgent := map[string]*AgentSkillAssignment{}
	for _, skill := range skills {
		for _, agent := range skill.Agents {
			item := byAgent[agent]
			if item == nil {
				item = &AgentSkillAssignment{Agent: agent}
				byAgent[agent] = item
			}
			item.Skills = append(item.Skills, skill.Name)
			if skill.Installed {
				item.Rules = append(item.Rules, fmt.Sprintf("Load `%s` when working on matching stack areas.", skill.Name))
			} else {
				item.Rules = append(item.Rules, fmt.Sprintf("Skill `%s` is recommended but missing; use baseline expertise and record the gap.", skill.Name))
			}
		}
	}
	var agents []AgentSkillAssignment
	for _, item := range byAgent {
		item.Skills = sortedUnique(item.Skills)
		item.Rules = sortedUnique(item.Rules)
		agents = append(agents, *item)
	}
	sort.Slice(agents, func(i, j int) bool { return agents[i].Agent < agents[j].Agent })
	return agents
}

func recommendedSkillInstalled(name string, registry *SkillRegistry) (bool, string) {
	if registry == nil {
		return false, ""
	}
	needle := normalizeSkillLookup(name)
	for _, skill := range registry.Skills {
		if normalizeSkillLookup(skill.Name) == needle || strings.Contains(normalizeSkillLookup(skill.Path), needle) || strings.Contains(needle, normalizeSkillLookup(skill.Name)) {
			return true, skill.Path
		}
	}
	return false, ""
}

func normalizeSkillLookup(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return value
}

func agentsForRecommendedSkill(skill string) []string {
	text := strings.ToLower(skill)
	agents := []string{"technical-lead"}
	if strings.Contains(text, "react") || strings.Contains(text, "next") || strings.Contains(text, "vue") || strings.Contains(text, "angular") || strings.Contains(text, "tailwind") || strings.Contains(text, "shadcn") || strings.Contains(text, "frontend") || strings.Contains(text, "responsive") || strings.Contains(text, "design") || strings.Contains(text, "interaction") || strings.Contains(text, "accessibility") || strings.Contains(text, "handoff") || strings.Contains(text, "openpencil") {
		agents = append(agents, "frontend-engineer", "ui-ux-designer")
	}
	if strings.Contains(text, "go") || strings.Contains(text, "python") || strings.Contains(text, "express") || strings.Contains(text, "hono") || strings.Contains(text, "nestjs") || strings.Contains(text, "fastify") || strings.Contains(text, "node") || strings.Contains(text, "prisma") || strings.Contains(text, "drizzle") || strings.Contains(text, "supabase") || strings.Contains(text, "postgres") || strings.Contains(text, "mysql") || strings.Contains(text, "mongo") || strings.Contains(text, "database") || strings.Contains(text, "backend") || strings.Contains(text, "api") {
		agents = append(agents, "backend-engineer")
	}
	if strings.Contains(text, "playwright") || strings.Contains(text, "testing") || strings.Contains(text, "vitest") || strings.Contains(text, "accessibility") || strings.Contains(text, "responsive") || strings.Contains(text, "qa") {
		agents = append(agents, "qa-security-reviewer")
	}
	return sortedUnique(agents)
}

func skillTechnologyRules() []skillTechRule {
	return []skillTechRule{
		{ID: "react", Name: "React", Kind: "frontend", StackNames: []string{"React"}, Skills: []string{"react-best-practices", "composition-patterns"}},
		{ID: "nextjs", Name: "Next.js", Kind: "frontend", StackNames: []string{"Next.js"}, Files: []string{"next.config.js", "next.config.mjs", "next.config.ts"}, Skills: []string{"next-best-practices", "next-cache-components", "next-upgrade"}},
		{ID: "typescript", Name: "TypeScript", Kind: "language", Packages: []string{"typescript"}, Files: []string{"tsconfig.json"}, Extensions: []string{".ts", ".tsx"}, StackNames: []string{"TypeScript"}, Skills: []string{"typescript-advanced-types"}},
		{ID: "tailwind", Name: "Tailwind CSS", Kind: "styling", StackNames: []string{"Tailwind CSS"}, Files: []string{"tailwind.config.js", "tailwind.config.ts", "tailwind.config.cjs"}, Skills: []string{"tailwind-css-patterns"}},
		{ID: "shadcn", Name: "shadcn/ui", Kind: "ui", StackNames: []string{"shadcn/ui"}, Files: []string{"components.json"}, Skills: []string{"shadcn"}},
		{ID: "vite", Name: "Vite", Kind: "tooling", StackNames: []string{"Vite"}, Files: []string{"vite.config.ts", "vite.config.js", "vite.config.mjs"}, Skills: []string{"vite-best-practices"}},
		{ID: "playwright", Name: "Playwright", Kind: "testing", StackNames: []string{"Playwright"}, Files: []string{"playwright.config.ts", "playwright.config.js"}, Skills: []string{"playwright-best-practices"}},
		{ID: "vitest", Name: "Vitest", Kind: "testing", StackNames: []string{"Vitest"}, Files: []string{"vitest.config.ts", "vitest.config.js"}, Skills: []string{"vitest-testing-patterns"}},
		{ID: "supabase", Name: "Supabase", Kind: "data", StackNames: []string{"Supabase"}, Skills: []string{"supabase-postgres-best-practices"}},
		{ID: "prisma", Name: "Prisma", Kind: "data", StackNames: []string{"Prisma"}, Files: []string{"prisma/schema.prisma"}, Skills: []string{"prisma-best-practices"}},
		{ID: "drizzle", Name: "Drizzle ORM", Kind: "data", StackNames: []string{"Drizzle ORM"}, Files: []string{"drizzle.config.ts"}, Skills: []string{"drizzle-orm-patterns"}},
		{ID: "zod", Name: "Zod", Kind: "validation", StackNames: []string{"Zod"}, Skills: []string{"zod"}},
		{ID: "react-hook-form", Name: "React Hook Form", Kind: "forms", StackNames: []string{"React Hook Form"}, Skills: []string{"react-hook-form"}},
		{ID: "nodejs", Name: "Node.js", Kind: "runtime", StackNames: []string{"Node.js"}, Skills: []string{"node-api-patterns"}},
		{ID: "express", Name: "Express", Kind: "backend", StackNames: []string{"Express"}, Skills: []string{"express-best-practices", "node-api-patterns"}},
		{ID: "nestjs", Name: "NestJS", Kind: "backend", StackNames: []string{"NestJS"}, Skills: []string{"nestjs-best-practices"}},
		{ID: "fastify", Name: "Fastify", Kind: "backend", StackNames: []string{"Fastify"}, Skills: []string{"fastify-best-practices", "node-api-patterns"}},
		{ID: "postgresql", Name: "PostgreSQL", Kind: "data", StackNames: []string{"PostgreSQL"}, Skills: []string{"postgres-best-practices"}},
		{ID: "mysql", Name: "MySQL", Kind: "data", StackNames: []string{"MySQL"}, Skills: []string{"mysql-best-practices"}},
		{ID: "mongodb", Name: "MongoDB", Kind: "data", StackNames: []string{"MongoDB"}, Skills: []string{"mongodb-best-practices"}},
		{ID: "go", Name: "Go", Kind: "language", Files: []string{"go.mod", "go.work"}, Extensions: []string{".go"}, StackNames: []string{"Go"}, Skills: []string{"go-testing", "go-best-practices"}},
		{ID: "python", Name: "Python", Kind: "language", Files: []string{"pyproject.toml", "requirements.txt", "uv.lock"}, Extensions: []string{".py"}, StackNames: []string{"Python"}, Skills: []string{"python-best-practices", "pytest-patterns"}},
		{ID: "docker", Name: "Docker", Kind: "platform", Files: []string{"Dockerfile", "docker-compose.yml", "docker-compose.yaml"}, StackNames: []string{"Docker"}, Skills: []string{"docker-best-practices"}},
	}
}

func skillComboRules() []skillComboRule {
	return []skillComboRule{
		{ID: "nextjs-supabase", Name: "Next.js + Supabase", Requires: []string{"nextjs", "supabase"}, Skills: []string{"supabase-nextjs-patterns"}},
		{ID: "nextjs-playwright", Name: "Next.js + Playwright", Requires: []string{"nextjs", "playwright"}, Skills: []string{"nextjs-e2e-testing"}},
		{ID: "react-shadcn", Name: "React + shadcn/ui", Requires: []string{"react", "shadcn"}, Skills: []string{"shadcn-react-patterns"}},
		{ID: "tailwind-shadcn", Name: "Tailwind CSS + shadcn/ui", Requires: []string{"tailwind", "shadcn"}, Skills: []string{"tailwind-shadcn-integration"}},
		{ID: "react-hook-form-zod", Name: "React Hook Form + Zod", Requires: []string{"react-hook-form", "zod"}, Skills: []string{"form-validation-patterns"}},
	}
}

func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
