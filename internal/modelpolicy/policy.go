package modelpolicy

import (
	"fmt"
	"sort"
	"strings"
)

const Version = "1"

const (
	TierFast      = "fast"
	TierBalanced  = "balanced"
	TierReasoning = "reasoning"
	TierDefault   = "default"
)

type Config struct {
	DefaultModel   string
	BalancedModel  string
	ReasoningModel string
	FastModel      string
	AgentModels    map[string]string
}

type Policy struct {
	Version     string            `json:"version"`
	GeneratedAt string            `json:"generated_at"`
	Tiers       map[string]string `json:"tiers"`
	Agents      map[string]string `json:"agents"`
	Notes       []string          `json:"notes,omitempty"`
}

func Default(cfg Config, generatedAt string) *Policy {
	cfg = normalizeConfig(cfg)
	return &Policy{
		Version:     Version,
		GeneratedAt: generatedAt,
		Tiers: map[string]string{
			TierFast:      cfg.FastModel,
			TierBalanced:  cfg.BalancedModel,
			TierReasoning: cfg.ReasoningModel,
			TierDefault:   cfg.DefaultModel,
		},
		Agents: map[string]string{
			"shipwright-orchestrator": TierBalanced,
			"project-manager":         TierFast,
			"product-owner":           TierBalanced,
			"technical-lead":          TierReasoning,
			"qa-security-reviewer":    TierReasoning,
			"ui-ux-designer":          TierBalanced,
			"frontend-engineer":       TierBalanced,
			"backend-engineer":        TierBalanced,
		},
		Notes: []string{
			"Orchestrator routes work and uses the balanced tier to keep delegation reliable.",
			"Reasoning is reserved for architecture, security review, and high-risk decisions.",
		},
	}
}

func Normalize(policy *Policy, fallback Config, generatedAt string) {
	if policy == nil {
		return
	}
	if policy.Version == "" {
		policy.Version = Version
	}
	if policy.GeneratedAt == "" {
		policy.GeneratedAt = generatedAt
	}
	if policy.Tiers == nil {
		policy.Tiers = map[string]string{}
	}
	if policy.Agents == nil {
		policy.Agents = map[string]string{}
	}
	defaults := Default(fallback, generatedAt)
	for tier, model := range defaults.Tiers {
		if strings.TrimSpace(policy.Tiers[tier]) == "" {
			policy.Tiers[tier] = model
		}
	}
	for agent, tier := range defaults.Agents {
		if strings.TrimSpace(policy.Agents[agent]) == "" {
			policy.Agents[agent] = tier
		}
	}
	if policy.Agents["shipwright-orchestrator"] == TierFast {
		policy.Agents["shipwright-orchestrator"] = TierBalanced
	}
	if policy.Agents["ui-ux-designer"] == "visual" {
		policy.Agents["ui-ux-designer"] = TierBalanced
	}
}

func Resolve(agent string, cfg Config, policy *Policy, legacyResolver func(string) string) string {
	cfg = normalizeConfig(cfg)
	if model := strings.TrimSpace(cfg.AgentModels[agent]); model != "" {
		return model
	}
	if policy != nil {
		Normalize(policy, cfg, "")
		if tier := strings.TrimSpace(policy.Agents[agent]); tier != "" {
			if model := strings.TrimSpace(policy.Tiers[tier]); model != "" {
				return model
			}
		}
	}
	if legacyResolver != nil {
		return legacyResolver(agent)
	}
	return cfg.DefaultModel
}

func RenderMarkdown(policy *Policy, generatedAt string) string {
	Normalize(policy, Config{}, generatedAt)
	var sb strings.Builder
	sb.WriteString("# Model Policy\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", policy.GeneratedAt))
	sb.WriteString("## Tiers\n\n")
	for _, tier := range []string{TierFast, TierBalanced, TierReasoning, TierDefault} {
		sb.WriteString(fmt.Sprintf("- `%s`: `%s`\n", tier, policy.Tiers[tier]))
	}
	sb.WriteString("\n## Agent bindings\n\n")
	agents := make([]string, 0, len(policy.Agents))
	for agent := range policy.Agents {
		agents = append(agents, agent)
	}
	sort.Strings(agents)
	for _, agent := range agents {
		sb.WriteString(fmt.Sprintf("- `%s` → `%s`\n", agent, policy.Agents[agent]))
	}
	sb.WriteString("\n## Rules\n\n")
	sb.WriteString("- Keep `shipwright-orchestrator` on the balanced tier; it routes work and must reliably delegate.\n")
	sb.WriteString("- Use reasoning for technical architecture, security, and high-risk reviews.\n")
	sb.WriteString("- Use per-agent overrides only for explicit project needs.\n")
	return sb.String()
}

func normalizeConfig(cfg Config) Config {
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = "anthropic/claude-sonnet-4-20250514"
	}
	if cfg.BalancedModel == "" {
		cfg.BalancedModel = cfg.DefaultModel
	}
	if cfg.ReasoningModel == "" {
		cfg.ReasoningModel = cfg.DefaultModel
	}
	if cfg.FastModel == "" {
		cfg.FastModel = cfg.DefaultModel
	}
	if cfg.AgentModels == nil {
		cfg.AgentModels = map[string]string{}
	}
	return cfg
}
