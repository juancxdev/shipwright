package harness

import agentsapp "shipwright/internal/agents/application"

type AgentStep = agentsapp.AgentStep
type Agent = agentsapp.Agent

func GetAgent(name string) *Agent {
	return agentsapp.GetAgent(name)
}

func AllAgents() []Agent {
	return agentsapp.AllAgents()
}

func ActiveAgentForPhase(phase string) *Agent {
	return agentsapp.ActiveAgentForPhase(phase)
}

func SecondaryAgentForPhase(phase string) *Agent {
	return agentsapp.SecondaryAgentForPhase(phase)
}
