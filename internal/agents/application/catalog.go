package application

import (
	"embed"
	"encoding/json"
	"fmt"

	agentdomain "shipwright/internal/agents/domain"
	lifecycle "shipwright/internal/lifecycle/domain"
)

//go:embed templates/agents.json
var agentCatalogFS embed.FS

type AgentStep = agentdomain.AgentStep
type Agent = agentdomain.Agent

var agentList = mustLoadAgentCatalog()

func mustLoadAgentCatalog() []Agent {
	data, err := agentCatalogFS.ReadFile("templates/agents.json")
	if err != nil {
		panic(fmt.Sprintf("cannot load agent catalog: %v", err))
	}
	var agents []Agent
	if err := json.Unmarshal(data, &agents); err != nil {
		panic(fmt.Sprintf("cannot parse agent catalog: %v", err))
	}
	return agents
}

func GetAgent(name string) *Agent {
	for i := range agentList {
		if agentList[i].Name == name {
			return &agentList[i]
		}
	}
	return nil
}

func AllAgents() []Agent {
	out := make([]Agent, len(agentList))
	copy(out, agentList)
	return out
}

func ActiveAgentForPhase(phase string) *Agent {
	switch phase {
	case lifecycle.StateDiscovery, lifecycle.StateProductContextReady, lifecycle.StateTechnicalScopeDraft, lifecycle.StateScopeReview:
		return GetAgent("product-owner")

	case lifecycle.StateScopeApproved, lifecycle.StateProjectPlanning:
		return GetAgent("project-manager")

	case lifecycle.StateUXDecision, lifecycle.StateUXDesign, lifecycle.StateUXApproval:
		return GetAgent("ui-ux-designer")

	case lifecycle.StateTechnicalDesign, lifecycle.StateBacklogReady, lifecycle.StateTechLeadReview:
		return GetAgent("technical-lead")

	case lifecycle.StateImplementation, lifecycle.StateIntegration:
		return GetAgent("frontend-engineer")

	case lifecycle.StateQASecurityReview:
		return GetAgent("qa-security-reviewer")

	case lifecycle.StateUserAcceptance:
		return GetAgent("project-manager")

	case lifecycle.StateChangeRequest:
		return GetAgent("project-manager")

	case lifecycle.StateClosed:
		return nil

	default:
		return nil
	}
}

func SecondaryAgentForPhase(phase string) *Agent {
	switch phase {
	case lifecycle.StateProductContextReady, lifecycle.StateTechnicalScopeDraft:
		return GetAgent("technical-lead")

	case lifecycle.StateTechnicalDesign:
		return GetAgent("frontend-engineer")

	case lifecycle.StateImplementation:
		return GetAgent("backend-engineer")

	default:
		return nil
	}
}
