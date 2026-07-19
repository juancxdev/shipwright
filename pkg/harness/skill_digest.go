package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

const SkillDigestsJSON = ".harness/skill-digests.json"
const SkillDigestsMarkdown = ".harness/skill-digests.md"
const SkillDigestsVersion = "1"

type SkillDigestSet struct {
	Version     string             `json:"version"`
	GeneratedAt string             `json:"generated_at"`
	Digests     []AgentSkillDigest `json:"digests"`
	Warnings    []string           `json:"warnings,omitempty"`
}

type AgentSkillDigest struct {
	Agent          string           `json:"agent"`
	RelevantSkills []DigestSkillRef `json:"relevant_skills"`
	CompactRules   []string         `json:"compact_rules"`
}

type DigestSkillRef struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

func RefreshSkillDigests() (*SkillDigestSet, error) {
	registry, err := LoadSkillRegistry()
	if err != nil {
		registry, err = BuildSkillRegistry()
		if err != nil {
			return nil, err
		}
	}
	profile, _ := LoadProjectProfile()
	digests := BuildSkillDigests(registry, profile)
	if err := SaveSkillDigests(digests); err != nil {
		return nil, err
	}
	return digests, nil
}

func RefreshSkillDigestsFromRegistry(registry *SkillRegistry) (*SkillDigestSet, error) {
	if registry == nil {
		return RefreshSkillDigests()
	}
	profile, _ := LoadProjectProfile()
	digests := BuildSkillDigests(registry, profile)
	if err := SaveSkillDigests(digests); err != nil {
		return nil, err
	}
	return digests, nil
}

func BuildSkillDigests(registry *SkillRegistry, profile *ProjectProfile) *SkillDigestSet {
	set := &SkillDigestSet{
		Version:     SkillDigestsVersion,
		GeneratedAt: NowISO(),
	}
	if registry == nil {
		set.Warnings = append(set.Warnings, "skill registry missing; no skill digests generated")
		return set
	}
	for _, agent := range digestAgentOrder() {
		digest := buildAgentSkillDigest(agent, registry, profile)
		set.Digests = append(set.Digests, digest)
	}
	if len(registry.Skills) == 0 {
		set.Warnings = append(set.Warnings, "skill registry has no skills")
	}
	return set
}

func SaveSkillDigests(digests *SkillDigestSet) error {
	if digests == nil {
		return fmt.Errorf("skill digests are nil")
	}
	data, err := json.MarshalIndent(digests, "", "  ")
	if err != nil {
		return err
	}
	if err := WriteFile(SkillDigestsJSON, string(data)+"\n"); err != nil {
		return err
	}
	return WriteFile(SkillDigestsMarkdown, RenderSkillDigestsMarkdown(digests))
}

func LoadSkillDigests() (*SkillDigestSet, error) {
	data, err := os.ReadFile(SkillDigestsJSON)
	if err != nil {
		return nil, err
	}
	var digests SkillDigestSet
	if err := json.Unmarshal(data, &digests); err != nil {
		return nil, err
	}
	return &digests, nil
}

func FindSkillDigest(digests *SkillDigestSet, agent string) *AgentSkillDigest {
	if digests == nil {
		return nil
	}
	needle := strings.ToLower(strings.TrimSpace(agent))
	for _, digest := range digests.Digests {
		if strings.ToLower(digest.Agent) == needle {
			copy := digest
			return &copy
		}
	}
	return nil
}

func RenderSkillDigestsMarkdown(digests *SkillDigestSet) string {
	var sb strings.Builder
	sb.WriteString("# Skill Digests\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n", digests.GeneratedAt))
	sb.WriteString(fmt.Sprintf("**Agents:** %d\n\n", len(digests.Digests)))

	for _, digest := range digests.Digests {
		sb.WriteString(fmt.Sprintf("## %s\n\n", digest.Agent))
		if len(digest.RelevantSkills) == 0 {
			sb.WriteString("No directly relevant skills detected. Use baseline Shipwright role instructions and record missing skill gaps if needed.\n\n")
		} else {
			sb.WriteString("### Relevant skills\n\n")
			for _, skill := range digest.RelevantSkills {
				sb.WriteString(fmt.Sprintf("- `%s` — %s (`%s`)\n", skill.Name, skill.Reason, skill.Path))
			}
			sb.WriteString("\n")
		}
		if len(digest.CompactRules) > 0 {
			sb.WriteString("### Compact rules\n\n")
			for _, rule := range digest.CompactRules {
				sb.WriteString(fmt.Sprintf("- %s\n", rule))
			}
			sb.WriteString("\n")
		}
	}

	if len(digests.Warnings) > 0 {
		sb.WriteString("## Warnings\n\n")
		for _, warning := range digests.Warnings {
			sb.WriteString(fmt.Sprintf("- %s\n", warning))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## How agents must use this\n\n")
	sb.WriteString("- Prefer this digest over loading every skill file.\n")
	sb.WriteString("- Load full skill files only when the digest says a skill is relevant and more detail is required.\n")
	sb.WriteString("- If no digest exists for your role, fall back to `.harness/skill-registry.md` and record the gap.\n")
	return sb.String()
}

func buildAgentSkillDigest(agent string, registry *SkillRegistry, profile *ProjectProfile) AgentSkillDigest {
	tags := digestTagsForAgent(agent)
	var refs []DigestSkillRef
	var rules []string
	for _, skill := range registry.Skills {
		if reason, ok := skillMatchesAgent(skill, agent, tags); ok {
			refs = append(refs, DigestSkillRef{Name: skill.Name, Path: skill.Path, Reason: reason})
			rule := compactRuleForSkill(skill)
			if rule != "" {
				rules = append(rules, rule)
			}
		}
	}
	rules = append(rules, profileCompactRulesForAgent(agent, profile)...)
	refs = uniqueDigestSkillRefs(refs)
	rules = sortedUnique(rules)
	if len(refs) > 8 {
		refs = refs[:8]
	}
	if len(rules) > 12 {
		rules = rules[:12]
	}
	return AgentSkillDigest{Agent: agent, RelevantSkills: refs, CompactRules: rules}
}

func skillMatchesAgent(skill SkillIndex, agent string, tags []string) (string, bool) {
	agentLower := strings.ToLower(agent)
	skillName := strings.ToLower(skill.Name)
	skillPath := strings.ToLower(skill.Path)
	if skillName == agentLower || strings.Contains(skillPath, agentLower) {
		return "direct role skill", true
	}
	for _, tag := range tags {
		if containsStringValue(skill.AppliesTo, tag) {
			return "matches " + tag + " capability", true
		}
	}
	text := strings.ToLower(skill.Name + " " + skill.Description + " " + strings.Join(skill.Triggers, " "))
	for _, tag := range tags {
		if strings.Contains(text, tag) {
			return "mentions " + tag, true
		}
	}
	return "", false
}

func compactRuleForSkill(skill SkillIndex) string {
	description := skill.Description
	if description == "" && len(skill.Triggers) > 0 {
		description = skill.Triggers[0]
	}
	if description == "" {
		return fmt.Sprintf("Use `%s` when its trigger matches; source `%s`.", skill.Name, skill.Path)
	}
	if len(description) > 180 {
		description = strings.TrimSpace(description[:177]) + "..."
	}
	return fmt.Sprintf("Use `%s` when relevant: %s", skill.Name, description)
}

func profileCompactRulesForAgent(agent string, profile *ProjectProfile) []string {
	if profile == nil {
		return nil
	}
	var rules []string
	if len(profile.Languages) > 0 {
		rules = append(rules, "Respect detected stack: "+strings.Join(profile.Languages, ", ")+".")
	}
	if len(profile.Commands.Test) > 0 && agentUsesTests(agent) {
		rules = append(rules, "Use detected test command for evidence: `"+profile.Commands.Test[0].Command+"`.")
	}
	if profile.TDD.RecommendedMode == "strict" && agentUsesTests(agent) {
		rules = append(rules, "Strict TDD is available; create/adjust failing tests before implementation when changing behavior.")
	}
	if profile.TDD.RecommendedMode != "strict" && agentUsesTests(agent) {
		rules = append(rules, "Do not claim strict TDD; project profile recommends `"+profile.TDD.RecommendedMode+"`.")
	}
	return rules
}

func digestAgentOrder() []string {
	agents := []string{"shipwright-orchestrator"}
	for _, skill := range AllAgentSkills() {
		agents = append(agents, skill.Name)
	}
	return agents
}

func digestTagsForAgent(agent string) []string {
	switch agent {
	case "shipwright-orchestrator":
		return []string{"testing", "frontend", "backend", "design", "docs", "go", "typescript"}
	case "product-owner", "project-manager":
		return []string{"docs"}
	case "technical-lead":
		return []string{"frontend", "backend", "testing", "go", "typescript", "docs"}
	case "ui-ux-designer":
		return []string{"design", "frontend", "docs"}
	case "frontend-engineer":
		return []string{"frontend", "typescript", "testing"}
	case "backend-engineer":
		return []string{"backend", "go", "testing"}
	case "qa-security-reviewer":
		return []string{"testing", "backend", "frontend"}
	default:
		return []string{"docs"}
	}
}

func agentUsesTests(agent string) bool {
	switch agent {
	case "technical-lead", "frontend-engineer", "backend-engineer", "qa-security-reviewer", "shipwright-orchestrator":
		return true
	default:
		return false
	}
}

func uniqueDigestSkillRefs(values []DigestSkillRef) []DigestSkillRef {
	seen := map[string]DigestSkillRef{}
	for _, value := range values {
		key := strings.ToLower(value.Name + ":" + value.Path)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = value
	}
	var out []DigestSkillRef
	for _, value := range seen {
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func containsStringValue(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
