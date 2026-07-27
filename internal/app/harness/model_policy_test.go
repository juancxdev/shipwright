package harness

import "testing"

func TestDefaultModelPolicyKeepsOrchestratorOnBalancedTier(t *testing.T) {
	cfg := DefaultOpenCodeExecutorConfig()
	cfg.FastModel = "fast/model"
	cfg.BalancedModel = "balanced/model"
	cfg.ReasoningModel = "reasoning/model"
	policy := DefaultModelPolicy(cfg)

	got := ResolveOpenCodeModelWithPolicy("shipwright-orchestrator", cfg, policy)
	if got != "balanced/model" {
		t.Fatalf("orchestrator model = %s, want balanced/model", got)
	}
	if tier := policy.Agents["shipwright-orchestrator"]; tier != ModelTierBalanced {
		t.Fatalf("orchestrator tier = %s, want %s", tier, ModelTierBalanced)
	}
	if _, ok := policy.Tiers[ModelTierDefault]; !ok {
		t.Fatal("default tier missing from model policy")
	}
}

func TestDefaultModelPolicyUsesReasoningForTechnicalLead(t *testing.T) {
	cfg := DefaultOpenCodeExecutorConfig()
	cfg.FastModel = "fast/model"
	cfg.ReasoningModel = "reasoning/model"
	policy := DefaultModelPolicy(cfg)

	got := ResolveOpenCodeModelWithPolicy("technical-lead", cfg, policy)
	if got != "reasoning/model" {
		t.Fatalf("technical-lead model = %s, want reasoning/model", got)
	}
}

func TestAgentModelOverrideWinsOverModelPolicy(t *testing.T) {
	cfg := DefaultOpenCodeExecutorConfig()
	cfg.FastModel = "fast/model"
	cfg.BalancedModel = "balanced/model"
	cfg.ReasoningModel = "reasoning/model"
	cfg.AgentModels["shipwright-orchestrator"] = "custom/router"
	policy := DefaultModelPolicy(cfg)

	got := ResolveOpenCodeModelWithPolicy("shipwright-orchestrator", cfg, policy)
	if got != "custom/router" {
		t.Fatalf("orchestrator model = %s, want custom/router", got)
	}
}
