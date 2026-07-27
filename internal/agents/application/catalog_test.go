package application

import (
	"strings"
	"testing"
)

func TestAgentCatalogCarriesSeniorRoleIdentity(t *testing.T) {
	cases := map[string]string{
		"product-owner":        "Senior Product Owner",
		"project-manager":      "Senior Delivery Manager",
		"technical-lead":       "Staff / Principal Technical Lead",
		"ui-ux-designer":       "Senior Product Designer",
		"frontend-engineer":    "Senior Frontend Engineer",
		"backend-engineer":     "Senior Backend Engineer",
		"qa-security-reviewer": "Senior QA / Security Reviewer",
	}

	for name, want := range cases {
		agent := GetAgent(name)
		if agent == nil {
			t.Fatalf("%s agent not loaded", name)
		}
		if !strings.Contains(agent.Purpose, want) {
			t.Fatalf("%s purpose missing %q:\n%s", name, want, agent.Purpose)
		}
	}
}
