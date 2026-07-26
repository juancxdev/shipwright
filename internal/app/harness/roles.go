package harness

import (
	"fmt"
	"os"
	"path/filepath"
)

func WriteRoles() error {
	sharedDir := filepath.Join(".harness/agents", "_shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		return fmt.Errorf("cannot create %s: %w", sharedDir, err)
	}

	commonPath := filepath.Join(sharedDir, "agent-common.md")
	if err := os.WriteFile(commonPath, []byte(AgentCommonProtocol), 0644); err != nil {
		return fmt.Errorf("cannot write %s: %w", commonPath, err)
	}

	for _, skill := range AllAgentSkills() {
		path := filepath.Join(".harness/agents", skill.Name+".md")
		if err := os.WriteFile(path, []byte(skill.Content), 0644); err != nil {
			return fmt.Errorf("cannot write %s: %w", path, err)
		}
	}

	return nil
}

func ListRoleNames() []string {
	skills := AllAgentSkills()
	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name + ".md"
	}
	return names
}
