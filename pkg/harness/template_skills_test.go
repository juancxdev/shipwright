package harness

import (
	"strings"
	"testing"
)

func TestAgentSkillsLoadFromProjectTemplates(t *testing.T) {
	skills := AllAgentSkills()
	if len(skills) != 7 {
		t.Fatalf("skills = %d, want 7", len(skills))
	}

	po := GetAgentSkill("product-owner")
	if po == nil {
		t.Fatal("product-owner skill not loaded")
	}
	if !strings.Contains(po.Content, "name: product-owner") {
		t.Fatalf("product-owner template content was not loaded")
	}
	if !strings.Contains(AgentCommonProtocol, "Shipwright Agent — Common Protocol") {
		t.Fatalf("shared agent common template was not loaded")
	}
}
