package harness

import (
	"encoding/json"
	"os"

	"shipwright/internal/modelpolicy"
)

const ModelPolicyJSON = ".harness/model-policy.json"
const ModelPolicyMarkdown = ".harness/model-policy.md"
const ModelPolicyVersion = modelpolicy.Version

const (
	ModelTierFast      = modelpolicy.TierFast
	ModelTierBalanced  = modelpolicy.TierBalanced
	ModelTierReasoning = modelpolicy.TierReasoning
	ModelTierVisual    = modelpolicy.TierVisual
)

type ModelPolicy modelpolicy.Policy

func DefaultModelPolicy(cfg PortableOpenCodeExecutorConfig) *ModelPolicy {
	return fromInternalModelPolicy(modelpolicy.Default(toModelPolicyConfig(cfg), NowISO()))
}

func SaveModelPolicy(policy *ModelPolicy) error {
	if policy == nil {
		return errNilModelPolicy()
	}
	modelpolicy.Normalize(policy.asInternal(), toModelPolicyConfig(DefaultOpenCodeExecutorConfig()), NowISO())
	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return err
	}
	if err := WriteFile(ModelPolicyJSON, string(data)+"\n"); err != nil {
		return err
	}
	return WriteFile(ModelPolicyMarkdown, RenderModelPolicyMarkdown(policy))
}

func LoadModelPolicy() (*ModelPolicy, error) {
	data, err := os.ReadFile(ModelPolicyJSON)
	if err != nil {
		return nil, err
	}
	var policy ModelPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, err
	}
	modelpolicy.Normalize(policy.asInternal(), toModelPolicyConfig(DefaultOpenCodeExecutorConfig()), NowISO())
	return &policy, nil
}

func LoadOrDefaultModelPolicy(cfg PortableOpenCodeExecutorConfig) *ModelPolicy {
	policy, err := LoadModelPolicy()
	if err == nil && policy != nil {
		return policy
	}
	return DefaultModelPolicy(cfg)
}

func (p *ModelPolicy) Normalize() {
	modelpolicy.Normalize(p.asInternal(), toModelPolicyConfig(DefaultOpenCodeExecutorConfig()), NowISO())
}

func ResolveOpenCodeModelWithPolicy(agent string, cfg PortableOpenCodeExecutorConfig, policy *ModelPolicy) string {
	legacy := func(agent string) string { return ResolveOpenCodeModel(agent, cfg) }
	return modelpolicy.Resolve(agent, toModelPolicyConfig(cfg), policy.asInternal(), legacy)
}

func RenderModelPolicyMarkdown(policy *ModelPolicy) string {
	return modelpolicy.RenderMarkdown(policy.asInternal(), NowISO())
}

func (p *ModelPolicy) asInternal() *modelpolicy.Policy {
	if p == nil {
		return nil
	}
	return (*modelpolicy.Policy)(p)
}

func fromInternalModelPolicy(policy *modelpolicy.Policy) *ModelPolicy {
	if policy == nil {
		return nil
	}
	return (*ModelPolicy)(policy)
}

func toModelPolicyConfig(cfg PortableOpenCodeExecutorConfig) modelpolicy.Config {
	cfg.Normalize()
	return modelpolicy.Config{
		DefaultModel:   cfg.DefaultModel,
		ReasoningModel: cfg.ReasoningModel,
		FastModel:      cfg.FastModel,
		AgentModels:    cfg.AgentModels,
	}
}

func errNilModelPolicy() error { return &modelPolicyError{"model policy is nil"} }

type modelPolicyError struct{ message string }

func (e *modelPolicyError) Error() string { return e.message }
