package harness

import "strings"

type OpenCodeModelOverrides struct {
	DefaultModel   string
	ReasoningModel string
	FastModel      string
	AgentModels    map[string]string
}

func ApplyOpenCodeModelOverrides(cfg *PortableConfig, overrides OpenCodeModelOverrides) bool {
	if cfg == nil {
		return false
	}
	changed := false
	cfg.Normalize()
	if value := trimEnv(overrides.DefaultModel); value != "" {
		cfg.Executors.OpenCode.DefaultModel = value
		changed = true
	}
	if value := trimEnv(overrides.ReasoningModel); value != "" {
		cfg.Executors.OpenCode.ReasoningModel = value
		changed = true
	}
	if value := trimEnv(overrides.FastModel); value != "" {
		cfg.Executors.OpenCode.FastModel = value
		changed = true
	}
	if len(overrides.AgentModels) > 0 {
		if cfg.Executors.OpenCode.AgentModels == nil {
			cfg.Executors.OpenCode.AgentModels = map[string]string{}
		}
		for agent, model := range overrides.AgentModels {
			agent = trimEnv(agent)
			model = trimEnv(model)
			if agent == "" || model == "" {
				continue
			}
			cfg.Executors.OpenCode.AgentModels[agent] = model
			changed = true
		}
	}
	cfg.Normalize()
	return changed
}

func ParseAgentModelOverrides(raw string) map[string]string {
	out := map[string]string{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		out[key] = value
	}
	return out
}

func ResolveOpenCodeModel(agent string, cfg PortableOpenCodeExecutorConfig) string {
	cfg.Normalize()
	if model := trimEnv(cfg.AgentModels[agent]); model != "" {
		return model
	}
	if isOpenCodeReasoningAgent(agent) {
		return cfg.ReasoningModel
	}
	if isOpenCodeFastAgent(agent) {
		return cfg.FastModel
	}
	return cfg.DefaultModel
}

func isOpenCodeReasoningAgent(agent string) bool {
	switch agent {
	case "shipwright-orchestrator", "technical-lead", "frontend-engineer", "backend-engineer", "qa-security-reviewer":
		return true
	default:
		return false
	}
}

func isOpenCodeFastAgent(agent string) bool {
	switch agent {
	case "product-owner", "project-manager", "ui-ux-designer":
		return true
	default:
		return false
	}
}

func KnownOpenCodeAgents() map[string]bool {
	agents := map[string]bool{"shipwright-orchestrator": true}
	for _, skill := range AllAgentSkills() {
		agents[skill.Name] = true
	}
	return agents
}
