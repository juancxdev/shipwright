package harness

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

const skillPackTemplateRoot = "templates/project/harness/skill-packs"

type SkillPack struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Kind        string             `json:"kind"`
	Description string             `json:"description,omitempty"`
	When        SkillPackCondition `json:"when"`
	Skills      []string           `json:"skills"`
	Agents      []string           `json:"agents,omitempty"`
}

type SkillPackCondition struct {
	Stacks            []string `json:"stacks,omitempty"`
	StackNameContains []string `json:"stack_name_contains,omitempty"`
	StackKindContains []string `json:"stack_kind_contains,omitempty"`
	Structure         []string `json:"structure,omitempty"`
	ArtifactsPrefix   []string `json:"artifacts_prefix,omitempty"`
}

var skillPacks []SkillPack

func init() {
	packs, err := loadSkillPacksFromTemplates()
	if err != nil {
		panic(fmt.Sprintf("cannot load skill pack templates: %v", err))
	}
	skillPacks = packs
}

func loadSkillPacksFromTemplates() ([]SkillPack, error) {
	entries, err := projectTemplateFS.ReadDir(skillPackTemplateRoot)
	if err != nil {
		return nil, err
	}
	packs := make([]SkillPack, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.ToSlash(filepath.Join(skillPackTemplateRoot, entry.Name()))
		data, err := projectTemplateFS.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var pack SkillPack
		if err := json.Unmarshal(data, &pack); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if strings.TrimSpace(pack.ID) == "" || strings.TrimSpace(pack.Name) == "" {
			return nil, fmt.Errorf("%s: skill pack requires id and name", path)
		}
		pack.Skills = sortedUnique(pack.Skills)
		pack.Agents = sortedUnique(pack.Agents)
		packs = append(packs, pack)
	}
	sort.Slice(packs, func(i, j int) bool { return packs[i].ID < packs[j].ID })
	return packs, nil
}

func AllSkillPacks() []SkillPack {
	out := make([]SkillPack, len(skillPacks))
	copy(out, skillPacks)
	return out
}

func MatchingSkillPacks(profile *ProjectProfile) []AssignedTechnology {
	if profile == nil {
		return nil
	}
	var matches []AssignedTechnology
	for _, pack := range skillPacks {
		if evidence := matchSkillPack(pack, profile); len(evidence) > 0 {
			matches = append(matches, AssignedTechnology{
				ID:       pack.ID,
				Name:     pack.Name,
				Kind:     pack.Kind,
				Evidence: sortedUnique(evidence),
				Skills:   sortedUnique(pack.Skills),
			})
		}
	}
	return matches
}

func matchSkillPack(pack SkillPack, profile *ProjectProfile) []string {
	var evidence []string
	if containsStringValue(pack.When.Structure, "frontend") && profile.Structure.HasFrontend {
		evidence = append(evidence, "profile-structure:frontend")
	}
	if containsStringValue(pack.When.Structure, "backend") && profile.Structure.HasBackend {
		evidence = append(evidence, "profile-structure:backend")
	}
	for _, artifact := range profile.ExistingArtifacts {
		normalized := filepath.ToSlash(artifact)
		for _, prefix := range pack.When.ArtifactsPrefix {
			if strings.HasPrefix(normalized, filepath.ToSlash(prefix)) {
				evidence = append(evidence, "profile-artifact:"+normalized)
			}
		}
	}
	for _, stack := range append(profile.Stacks, profile.PlannedStacks...) {
		if skillPackStackMatches(pack.When, stack) {
			evidence = append(evidence, "profile-stack:"+stack.Name)
		}
	}
	return sortedUnique(evidence)
}

func skillPackStackMatches(condition SkillPackCondition, stack StackSignal) bool {
	stackName := strings.ToLower(stack.Name)
	stackKind := strings.ToLower(stack.Kind)
	for _, expected := range condition.Stacks {
		if stackName == strings.ToLower(expected) {
			return true
		}
	}
	for _, needle := range condition.StackNameContains {
		if strings.Contains(stackName, strings.ToLower(needle)) {
			return true
		}
	}
	for _, needle := range condition.StackKindContains {
		if strings.Contains(stackKind, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}
