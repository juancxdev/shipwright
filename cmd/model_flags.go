package cmd

import (
	"fmt"
	"strings"

	"shipwright/pkg/harness"
)

type openCodeModelFlagParseResult struct {
	Overrides harness.OpenCodeModelOverrides
	Used      bool
}

func parseOpenCodeModelFlags(args []string) openCodeModelFlagParseResult {
	result := openCodeModelFlagParseResult{Overrides: harness.OpenCodeModelOverrides{AgentModels: map[string]string{}}}
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		switch {
		case arg == "--opencode-default-model" || arg == "--default-model":
			value, next := requiredFlagValue(args, i, arg)
			result.Overrides.DefaultModel = value
			result.Used = true
			i = next
		case strings.HasPrefix(arg, "--opencode-default-model="):
			result.Overrides.DefaultModel = strings.TrimPrefix(arg, "--opencode-default-model=")
			result.Used = true
		case strings.HasPrefix(arg, "--default-model="):
			result.Overrides.DefaultModel = strings.TrimPrefix(arg, "--default-model=")
			result.Used = true
		case arg == "--opencode-reasoning-model" || arg == "--reasoning-model":
			value, next := requiredFlagValue(args, i, arg)
			result.Overrides.ReasoningModel = value
			result.Used = true
			i = next
		case strings.HasPrefix(arg, "--opencode-reasoning-model="):
			result.Overrides.ReasoningModel = strings.TrimPrefix(arg, "--opencode-reasoning-model=")
			result.Used = true
		case strings.HasPrefix(arg, "--reasoning-model="):
			result.Overrides.ReasoningModel = strings.TrimPrefix(arg, "--reasoning-model=")
			result.Used = true
		case arg == "--opencode-fast-model" || arg == "--fast-model":
			value, next := requiredFlagValue(args, i, arg)
			result.Overrides.FastModel = value
			result.Used = true
			i = next
		case strings.HasPrefix(arg, "--opencode-fast-model="):
			result.Overrides.FastModel = strings.TrimPrefix(arg, "--opencode-fast-model=")
			result.Used = true
		case strings.HasPrefix(arg, "--fast-model="):
			result.Overrides.FastModel = strings.TrimPrefix(arg, "--fast-model=")
			result.Used = true
		case arg == "--opencode-agent-model" || arg == "--agent-model":
			value, next := requiredFlagValue(args, i, arg)
			mergeAgentModelFlag(result.Overrides.AgentModels, value)
			result.Used = true
			i = next
		case strings.HasPrefix(arg, "--opencode-agent-model="):
			mergeAgentModelFlag(result.Overrides.AgentModels, strings.TrimPrefix(arg, "--opencode-agent-model="))
			result.Used = true
		case strings.HasPrefix(arg, "--agent-model="):
			mergeAgentModelFlag(result.Overrides.AgentModels, strings.TrimPrefix(arg, "--agent-model="))
			result.Used = true
		}
	}
	return result
}

func requiredFlagValue(args []string, index int, flag string) (string, int) {
	if index+1 >= len(args) || strings.HasPrefix(args[index+1], "--") {
		Fail(fmt.Sprintf("missing value for %s", flag))
	}
	return args[index+1], index + 1
}

func mergeAgentModelFlag(target map[string]string, raw string) {
	for agent, model := range harness.ParseAgentModelOverrides(raw) {
		target[agent] = model
	}
}

func persistOpenCodeModelOverrides(overrides harness.OpenCodeModelOverrides) bool {
	cfg, err := harness.LoadPortableConfig()
	if err != nil {
		Fail(fmt.Sprintf("cannot load portable config for OpenCode model overrides: %s", err))
	}
	changed := harness.ApplyOpenCodeModelOverrides(cfg, overrides)
	if changed {
		if err := cfg.Save(); err != nil {
			Fail(fmt.Sprintf("cannot save OpenCode model config: %s", err))
		}
	}
	return changed
}
