package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const SkillRegistryJSON = ".harness/skill-registry.json"
const SkillRegistryMarkdown = ".harness/skill-registry.md"
const SkillRegistryVersion = "1"

type SkillRegistry struct {
	Version     string       `json:"version"`
	GeneratedAt string       `json:"generated_at"`
	Root        string       `json:"root"`
	Skills      []SkillIndex `json:"skills"`
	Sources     []string     `json:"sources"`
	Warnings    []string     `json:"warnings,omitempty"`
}

type SkillIndex struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Source      string   `json:"source"`
	Path        string   `json:"path"`
	Triggers    []string `json:"triggers,omitempty"`
	AppliesTo   []string `json:"applies_to,omitempty"`
}

type skillScanSource struct {
	Root   string
	Source string
}

func RefreshSkillRegistry() (*SkillRegistry, error) {
	registry, err := BuildSkillRegistry()
	if err != nil {
		return nil, err
	}
	if err := SaveSkillRegistry(registry); err != nil {
		return nil, err
	}
	return registry, nil
}

func BuildSkillRegistry() (*SkillRegistry, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	registry := &SkillRegistry{
		Version:     SkillRegistryVersion,
		GeneratedAt: NowISO(),
		Root:        root,
	}

	for _, source := range skillScanSources() {
		if !dirExists(source.Root) {
			continue
		}
		registry.Sources = append(registry.Sources, source.Root)
		if err := scanSkillSource(registry, source); err != nil {
			registry.Warnings = append(registry.Warnings, fmt.Sprintf("%s: %s", source.Root, err))
		}
	}

	registry.Skills = uniqueSkillIndexes(registry.Skills)
	registry.Sources = sortedUnique(registry.Sources)
	if len(registry.Skills) == 0 {
		registry.Warnings = append(registry.Warnings, "no SKILL.md files found in known skill locations")
	}
	return registry, nil
}

func SaveSkillRegistry(registry *SkillRegistry) error {
	if registry == nil {
		return fmt.Errorf("skill registry is nil")
	}
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}
	if err := WriteFile(SkillRegistryJSON, string(data)+"\n"); err != nil {
		return err
	}
	return WriteFile(SkillRegistryMarkdown, RenderSkillRegistryMarkdown(registry))
}

func LoadSkillRegistry() (*SkillRegistry, error) {
	data, err := os.ReadFile(SkillRegistryJSON)
	if err != nil {
		return nil, err
	}
	var registry SkillRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, err
	}
	return &registry, nil
}

func RenderSkillRegistryMarkdown(registry *SkillRegistry) string {
	var sb strings.Builder
	sb.WriteString("# Skill Registry\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n", registry.GeneratedAt))
	sb.WriteString(fmt.Sprintf("**Skills indexed:** %d\n\n", len(registry.Skills)))

	if len(registry.Sources) > 0 {
		sb.WriteString("## Sources\n\n")
		for _, source := range registry.Sources {
			sb.WriteString(fmt.Sprintf("- `%s`\n", source))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Indexed skills\n\n")
	if len(registry.Skills) == 0 {
		sb.WriteString("No skills detected. Run `shipwright skills refresh` after generating executor assets or adding project skills.\n")
	} else {
		for _, skill := range registry.Skills {
			sb.WriteString(fmt.Sprintf("### %s\n\n", skill.Name))
			sb.WriteString(fmt.Sprintf("- Source: `%s`\n", skill.Source))
			sb.WriteString(fmt.Sprintf("- Path: `%s`\n", skill.Path))
			if skill.Description != "" {
				sb.WriteString(fmt.Sprintf("- Description: %s\n", skill.Description))
			}
			if len(skill.Triggers) > 0 {
				sb.WriteString(fmt.Sprintf("- Triggers: %s\n", strings.Join(skill.Triggers, "; ")))
			}
			if len(skill.AppliesTo) > 0 {
				sb.WriteString(fmt.Sprintf("- Applies to: `%s`\n", strings.Join(skill.AppliesTo, ", ")))
			}
			sb.WriteString("\n")
		}
	}

	if len(registry.Warnings) > 0 {
		sb.WriteString("## Warnings\n\n")
		for _, warning := range registry.Warnings {
			sb.WriteString(fmt.Sprintf("- %s\n", warning))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## How agents must use this\n\n")
	sb.WriteString("- Read this registry before delegating specialized work.\n")
	sb.WriteString("- Load only relevant skill instructions; do not flood subagents with every skill.\n")
	sb.WriteString("- If a needed skill is missing, continue with explicit fallback and record the gap.\n")
	sb.WriteString("- Refresh this registry after adding, removing, or regenerating skills.\n")
	return sb.String()
}

func FindSkill(registry *SkillRegistry, name string) *SkillIndex {
	if registry == nil {
		return nil
	}
	needle := strings.ToLower(strings.TrimSpace(name))
	for _, skill := range registry.Skills {
		if strings.ToLower(skill.Name) == needle {
			copy := skill
			return &copy
		}
	}
	return nil
}

func skillScanSources() []skillScanSource {
	return []skillScanSource{
		{Root: filepath.Join(".opencode", "skills"), Source: "opencode"},
		{Root: filepath.Join(".harness", "agents"), Source: "shipwright-agents"},
		{Root: filepath.Join(".agent", "skills"), Source: "project-agent"},
		{Root: "skills", Source: "project"},
	}
}

func scanSkillSource(registry *SkillRegistry, source skillScanSource) error {
	return filepath.WalkDir(source.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "SKILL.md" && !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		index, err := parseSkillFile(path, source.Source)
		if err != nil {
			registry.Warnings = append(registry.Warnings, fmt.Sprintf("cannot parse %s: %s", path, err))
			return nil
		}
		if index.Name != "" {
			registry.Skills = append(registry.Skills, index)
		}
		return nil
	})
}

func parseSkillFile(path, source string) (SkillIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SkillIndex{}, err
	}
	content := string(data)
	frontmatter := parseFrontmatter(content)
	name := strings.TrimSpace(frontmatter["name"])
	if name == "" {
		name = inferSkillName(path)
	}
	description := strings.TrimSpace(frontmatter["description"])
	if description == "" {
		description = firstMarkdownParagraph(content)
	}
	return SkillIndex{
		Name:        name,
		Description: normalizeWhitespace(description),
		Source:      source,
		Path:        filepath.ToSlash(path),
		Triggers:    extractTriggers(content),
		AppliesTo:   inferSkillAppliesTo(name, content),
	}, nil
}

func parseFrontmatter(content string) map[string]string {
	result := map[string]string{}
	content = strings.TrimPrefix(content, "\ufeff")
	if !strings.HasPrefix(content, "---\n") {
		return result
	}
	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return result
	}
	block := content[4 : 4+end]
	var currentKey string
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if currentKey != "" {
				result[currentKey] = strings.TrimSpace(result[currentKey] + " " + strings.TrimSpace(line))
			}
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		currentKey = strings.TrimSpace(parts[0])
		result[currentKey] = strings.Trim(strings.TrimSpace(parts[1]), `"'`)
	}
	return result
}

func inferSkillName(path string) string {
	dir := filepath.Base(filepath.Dir(path))
	if strings.EqualFold(filepath.Base(path), "SKILL.md") && dir != "." && dir != string(filepath.Separator) {
		return dir
	}
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func firstMarkdownParagraph(content string) string {
	content = stripFrontmatter(content)
	for _, block := range strings.Split(content, "\n\n") {
		block = strings.TrimSpace(block)
		if block == "" || strings.HasPrefix(block, "#") || strings.HasPrefix(block, "---") {
			continue
		}
		return block
	}
	return ""
}

func stripFrontmatter(content string) string {
	content = strings.TrimPrefix(content, "\ufeff")
	if !strings.HasPrefix(content, "---\n") {
		return content
	}
	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return content
	}
	return content[4+end+4:]
}

func extractTriggers(content string) []string {
	var triggers []string
	lowerLines := strings.Split(content, "\n")
	for _, line := range lowerLines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "trigger:") || strings.HasPrefix(lower, "triggers:") || strings.Contains(lower, "trigger rules") || strings.Contains(lower, "when to use") {
			triggers = append(triggers, normalizeWhitespace(trimmed))
		}
		if len(triggers) >= 8 {
			break
		}
	}
	return sortedUnique(triggers)
}

func inferSkillAppliesTo(name, content string) []string {
	lower := strings.ToLower(name + "\n" + content)
	var applies []string
	checks := map[string][]string{
		"go":         {"go ", "golang", "go test", ".go"},
		"typescript": {"typescript", "tsx", "tsconfig", ".ts"},
		"frontend":   {"frontend", "react", "vue", "angular", "css", "ui", "ux"},
		"backend":    {"backend", "api", "database", "server"},
		"testing":    {"test", "tdd", "verify", "qa"},
		"design":     {"design", "figma", "openpencil", "wireframe", "prototype"},
		"docs":       {"documentation", "readme", "docs"},
	}
	for label, needles := range checks {
		for _, needle := range needles {
			if strings.Contains(lower, needle) {
				applies = append(applies, label)
				break
			}
		}
	}
	return sortedUnique(applies)
}

func uniqueSkillIndexes(values []SkillIndex) []SkillIndex {
	seen := map[string]SkillIndex{}
	for _, value := range values {
		key := strings.ToLower(value.Source + ":" + value.Name + ":" + value.Path)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = value
	}
	var out []SkillIndex
	for _, value := range seen {
		value.Triggers = sortedUnique(value.Triggers)
		value.AppliesTo = sortedUnique(value.AppliesTo)
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source == out[j].Source {
			return out[i].Name < out[j].Name
		}
		return out[i].Source < out[j].Source
	})
	return out
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
