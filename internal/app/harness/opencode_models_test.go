package harness

import "testing"

func TestApplyOpenCodeModelOverridesDefaultsBalancedToFastWhenDefaultImplicit(t *testing.T) {
	cfg := DefaultPortableConfig()

	changed := ApplyOpenCodeModelOverrides(cfg, OpenCodeModelOverrides{
		FastModel:      "opencode-go/deepseek-v4-flash",
		ReasoningModel: "opencode-go/deepseek-v4-flash",
	})

	if !changed {
		t.Fatal("ApplyOpenCodeModelOverrides changed = false, want true")
	}
	if cfg.Executors.OpenCode.DefaultModel != "opencode-go/deepseek-v4-flash" {
		t.Fatalf("default model = %q, want fast model", cfg.Executors.OpenCode.DefaultModel)
	}
	if cfg.Executors.OpenCode.BalancedModel != "opencode-go/deepseek-v4-flash" {
		t.Fatalf("balanced model = %q, want fast model", cfg.Executors.OpenCode.BalancedModel)
	}
	if cfg.Executors.OpenCode.FastModel != "opencode-go/deepseek-v4-flash" {
		t.Fatalf("fast model = %q", cfg.Executors.OpenCode.FastModel)
	}
	if cfg.Executors.OpenCode.ReasoningModel != "opencode-go/deepseek-v4-flash" {
		t.Fatalf("reasoning model = %q", cfg.Executors.OpenCode.ReasoningModel)
	}
}

func TestApplyOpenCodeModelOverridesDoesNotOverwriteCustomDefault(t *testing.T) {
	cfg := DefaultPortableConfig()
	cfg.Executors.OpenCode.DefaultModel = "custom/balanced"

	ApplyOpenCodeModelOverrides(cfg, OpenCodeModelOverrides{
		FastModel: "opencode-go/deepseek-v4-flash",
	})

	if cfg.Executors.OpenCode.DefaultModel != "custom/balanced" {
		t.Fatalf("default model = %q, want custom/balanced", cfg.Executors.OpenCode.DefaultModel)
	}
	if cfg.Executors.OpenCode.BalancedModel != "custom/balanced" {
		t.Fatalf("balanced model = %q, want custom/balanced", cfg.Executors.OpenCode.BalancedModel)
	}
	if cfg.Executors.OpenCode.FastModel != "opencode-go/deepseek-v4-flash" {
		t.Fatalf("fast model = %q", cfg.Executors.OpenCode.FastModel)
	}
}

func TestApplyOpenCodeModelOverridesSupportsExplicitBalancedModel(t *testing.T) {
	cfg := DefaultPortableConfig()

	ApplyOpenCodeModelOverrides(cfg, OpenCodeModelOverrides{
		BalancedModel: "custom/mid",
		FastModel:     "custom/fast",
	})

	if cfg.Executors.OpenCode.BalancedModel != "custom/mid" {
		t.Fatalf("balanced model = %q, want custom/mid", cfg.Executors.OpenCode.BalancedModel)
	}
	if cfg.Executors.OpenCode.FastModel != "custom/fast" {
		t.Fatalf("fast model = %q, want custom/fast", cfg.Executors.OpenCode.FastModel)
	}
}
