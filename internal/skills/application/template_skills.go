package application

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// projectTemplateFS contains static project files that Shipwright copies into a
// user project during init/executor generation. Keep large prompts, skills, and
// markdown scaffolding here instead of hardcoding them in Go.
//
//go:embed templates/project/harness/agents/*.md templates/project/harness/agents/_shared/*.md templates/project/harness/skills/*/SKILL.md templates/project/harness/skill-packs/*.json
var projectTemplateFS embed.FS

const agentTemplateRoot = "templates/project/harness/agents"
const curatedSkillTemplateRoot = "templates/project/harness/skills"
const AgentCommonFilename = "_shared/agent-common.md"

type AgentSkill struct {
	Name    string
	Content string
}

var (
	AgentCommonProtocol string
	agentSkills         []AgentSkill
	curatedSkills       []AgentSkill
)

func init() {
	common, err := readProjectTemplate(filepath.ToSlash(filepath.Join(agentTemplateRoot, AgentCommonFilename)))
	if err != nil {
		panic(fmt.Sprintf("cannot load agent common template: %v", err))
	}
	AgentCommonProtocol = common

	skills, err := loadAgentSkillsFromTemplates()
	if err != nil {
		panic(fmt.Sprintf("cannot load agent skill templates: %v", err))
	}
	agentSkills = skills

	curated, err := loadCuratedSkillsFromTemplates()
	if err != nil {
		panic(fmt.Sprintf("cannot load curated skill templates: %v", err))
	}
	curatedSkills = curated
}

func readProjectTemplate(path string) (string, error) {
	data, err := projectTemplateFS.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func loadAgentSkillsFromTemplates() ([]AgentSkill, error) {
	entries, err := projectTemplateFS.ReadDir(agentTemplateRoot)
	if err != nil {
		return nil, err
	}

	skills := make([]AgentSkill, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") || entry.Name() == ".gitkeep" {
			continue
		}

		path := filepath.ToSlash(filepath.Join(agentTemplateRoot, entry.Name()))
		content, err := readProjectTemplate(path)
		if err != nil {
			return nil, err
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		skills = append(skills, AgentSkill{Name: name, Content: content})
	}

	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})
	return skills, nil
}

func loadCuratedSkillsFromTemplates() ([]AgentSkill, error) {
	entries, err := projectTemplateFS.ReadDir(curatedSkillTemplateRoot)
	if err != nil {
		return nil, err
	}

	skills := make([]AgentSkill, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".gitkeep" {
			continue
		}

		path := filepath.ToSlash(filepath.Join(curatedSkillTemplateRoot, entry.Name(), "SKILL.md"))
		content, err := readProjectTemplate(path)
		if err != nil {
			return nil, err
		}
		skills = append(skills, AgentSkill{Name: entry.Name(), Content: content})
	}

	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})
	return skills, nil
}

func GetAgentSkill(name string) *AgentSkill {
	for i := range agentSkills {
		if agentSkills[i].Name == name {
			return &agentSkills[i]
		}
	}
	return nil
}

func AllAgentSkills() []AgentSkill {
	out := make([]AgentSkill, len(agentSkills))
	copy(out, agentSkills)
	return out
}

func GetCuratedSkill(name string) *AgentSkill {
	for i := range curatedSkills {
		if curatedSkills[i].Name == name {
			return &curatedSkills[i]
		}
	}
	return nil
}

func AllCuratedSkills() []AgentSkill {
	out := make([]AgentSkill, len(curatedSkills))
	copy(out, curatedSkills)
	return out
}
