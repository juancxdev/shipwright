package harness

import agentdomain "shipwright/internal/agents/domain"

type AgentStep = agentdomain.AgentStep
type Agent = agentdomain.Agent

func GetAgent(name string) *Agent {
	return agentdomain.GetAgent(name)
}

func AllAgents() []Agent {
	return agentdomain.AllAgents()
}

func ActiveAgentForPhase(phase string) *Agent {
	return agentdomain.ActiveAgentForPhase(phase)
}

func SecondaryAgentForPhase(phase string) *Agent {
	return agentdomain.SecondaryAgentForPhase(phase)
}
