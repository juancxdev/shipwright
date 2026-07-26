package harness

import "testing"

func TestDefaultModelPolicyKeepsOrchestratorOnFastTier(t *testing.T) {
	cfg := DefaultOpenCodeExecutorConfig()
	cfg.FastModel = "fast/model"
	cfg.ReasoningModel = "reasoning/model"
	policy := DefaultModelPolicy(cfg)

	got := ResolveOpenCodeModelWithPolicy("shipwright-orchestrator", cfg, policy)
	if got != "fast/model" {
		t.Fatalf("orchestrator model = %s, want fast/model", got)
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
	cfg.ReasoningModel = "reasoning/model"
	cfg.AgentModels["shipwright-orchestrator"] = "custom/router"
	policy := DefaultModelPolicy(cfg)

	got := ResolveOpenCodeModelWithPolicy("shipwright-orchestrator", cfg, policy)
	if got != "custom/router" {
		t.Fatalf("orchestrator model = %s, want custom/router", got)
	}
}
