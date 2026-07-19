package cmd

import (
	"fmt"
	"strings"

	"shipwright/pkg/harness"
)

func Executor(args []string) {
	EnsureHarness()
	if len(args) == 0 {
		PrintUsage()
		Fail("usage: shipwright executor <list|status|generate>")
	}

	switch args[0] {
	case "list":
		executorList(args[1:])
	case "status":
		executorStatus(args[1:])
	case "generate":
		executorGenerate(args[1:])
	default:
		Fail(fmt.Sprintf("unknown executor subcommand: %s\n\nValid: list | status | generate", args[0]))
	}
}

func executorList(args []string) {
	fmt.Println("Shipwright — Executor Adapters")
	fmt.Println("=================================")
	fmt.Println()
	for _, name := range harness.ExecutorNames() {
		adapter, _ := harness.GetExecutorAdapter(name)
		fmt.Printf("  %-10s %s\n", adapter.Name(), adapter.Description())
	}
}

func executorStatus(args []string) {
	name := selectedExecutorName(args)
	adapter, err := harness.GetExecutorAdapter(name)
	if err != nil {
		Fail(err.Error())
	}
	status, err := adapter.Status()
	if err != nil {
		Fail(fmt.Sprintf("cannot inspect executor: %s", err))
	}

	fmt.Println("Shipwright — Executor Status")
	fmt.Println("==============================")
	fmt.Println()
	fmt.Printf("Executor:   %s\n", status.Name)
	fmt.Printf("Configured: %s\n", boolYesNo(status.Configured))
	fmt.Println()
	if len(status.Missing) > 0 {
		fmt.Println("Missing files:")
		for _, file := range status.Missing {
			fmt.Printf("  - %s\n", file)
		}
		fmt.Println()
	}
	if len(status.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, warning := range status.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
		fmt.Println()
	}
	fmt.Printf("Tracked files: %d\n", len(status.Files))
	if !status.Configured {
		fmt.Printf("Run: shipwright executor generate %s\n", status.Name)
	}
}

func executorGenerate(args []string) {
	if len(args) == 0 {
		Fail("usage: shipwright executor generate <generic|opencode> [--reasoning-model <model>] [--fast-model <model>] [--agent-model role=model]")
	}
	name := strings.TrimSpace(args[0])
	modelFlags := parseOpenCodeModelFlags(args[1:])
	if name == harness.ExecutorOpenCode && modelFlags.Used {
		if persistOpenCodeModelOverrides(modelFlags.Overrides) {
			PrintSuccess("OpenCode model config actualizada (.harness/config.json)")
		}
	}
	if policy, err := harness.RefreshTDDPolicy(); err == nil {
		PrintSuccess(fmt.Sprintf("TDD policy actualizada (%s)", policy.Mode))
	} else {
		PrintInfo(fmt.Sprintf("TDD policy no actualizada: %s", err))
	}

	result, err := harness.GenerateExecutor(name)
	if err != nil {
		Fail(err.Error())
	}
	PrintSuccess(result.Message)
	if registry, err := harness.RefreshSkillRegistry(); err == nil {
		PrintSuccess(fmt.Sprintf("Skill registry actualizado (%d skills)", len(registry.Skills)))
		if digests, err := harness.RefreshSkillDigestsFromRegistry(registry); err == nil {
			PrintSuccess(fmt.Sprintf("Skill digests actualizados (%d agentes)", len(digests.Digests)))
		} else {
			PrintInfo(fmt.Sprintf("skill digests no actualizados: %s", err))
		}
	} else {
		PrintInfo(fmt.Sprintf("skill registry no actualizado: %s", err))
	}
	fmt.Printf("Created: %d, updated: %d\n", len(result.FilesCreated), len(result.FilesUpdated))
	if len(result.FilesCreated) > 0 {
		fmt.Println("Created files:")
		for _, file := range result.FilesCreated {
			fmt.Printf("  + %s\n", file)
		}
	}
	if len(result.FilesUpdated) > 0 {
		fmt.Println("Updated files:")
		for _, file := range result.FilesUpdated {
			fmt.Printf("  ~ %s\n", file)
		}
	}
}

func selectedExecutorName(args []string) string {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return args[0]
	}
	state, err := harness.LoadExecutorState()
	if err == nil && state.Executor != "" {
		return state.Executor
	}
	return harness.ExecutorGeneric
}
