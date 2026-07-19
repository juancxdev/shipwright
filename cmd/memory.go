package cmd

import (
	"encoding/json"
	"fmt"

	"shipwright/pkg/harness"
)

func Memory(args []string) {
	EnsureHarness()

	if len(args) == 0 {
		PrintUsage()
		Fail("usage: shipwright memory <status|enable|disable|flush|mark-synced>")
	}

	subcommand := args[0]
	rest := args[1:]

	switch subcommand {
	case "status":
		memoryStatus(rest)
	case "enable":
		memoryEnable(rest)
	case "disable":
		memoryDisable(rest)
	case "flush":
		memoryFlush(rest)
	case "mark-synced":
		memoryMarkSynced(rest)
	default:
		Fail(fmt.Sprintf("unknown memory subcommand: %s\n\nValid: status | enable | disable | flush | mark-synced", subcommand))
	}
}

func memoryStatus(args []string) {
	integrations, err := harness.LoadIntegrations()
	if err != nil {
		Fail(fmt.Sprintf("cannot load integrations: %s", err))
	}

	fmt.Println("Shipwright — Memory Status")
	fmt.Println("================================")
	fmt.Println()

	engramOn := integrations.IsEngramEnabled()
	mode := "fallback (progress/decisions.md)"
	if engramOn {
		mode = "engram (.harness/memory-queue.json)"
	}
	fmt.Printf("Adapter:  %s\n", mode)
	fmt.Printf("Engram:   %s (%s)\n", boolEnabled(engramOn), integrations.Engram.Status)
	fmt.Printf("Fallback: %s\n", integrations.Engram.Fallback)
	fmt.Println()

	if engramOn {
		adapter := harness.NewEngramMemoryAdapter()
		total, pending, synced := adapter.Stats()

		fmt.Printf("Queue:     .harness/memory-queue.json\n")
		fmt.Printf("  Total:   %d\n", total)
		fmt.Printf("  Pending: %d\n", pending)
		fmt.Printf("  Synced:  %d\n", synced)
		fmt.Println()

		pendingEvents := adapter.PendingEvents()
		if len(pendingEvents) > 0 {
			fmt.Println("Pending events:")
			for _, ev := range pendingEvents {
				fmt.Printf("  [%s] %s — %s\n", ev.Type, ev.ID, ev.Title)
			}
			fmt.Println()
			fmt.Println("To sync: shipwright memory flush")
			fmt.Println("After AI sync: shipwright memory mark-synced")
		} else {
			fmt.Println("No pending events.")
		}
	} else {
		localCount := harness.CountLocalEntries()
		fmt.Printf("Local log: %s\n", harness.DecisionsFile)
		fmt.Printf("  Entries: %d\n", localCount)
		fmt.Println()
		fmt.Println("To enable Engram: shipwright memory enable")
	}
}

func memoryEnable(args []string) {
	integrations, err := harness.LoadIntegrations()
	if err != nil {
		Fail(fmt.Sprintf("cannot load integrations: %s", err))
	}

	if integrations.IsEngramEnabled() {
		PrintInfo("Engram already enabled.")
		return
	}

	integrations.EnableEngram()
	if err := integrations.Save(); err != nil {
		Fail(fmt.Sprintf("cannot save integrations: %s", err))
	}

	PrintSuccess("Engram memory enabled.")
	PrintInfo("Future memory events will be queued in .harness/memory-queue.json")
	PrintInfo("The AI agent should run 'shipwright memory flush' and call mem_save for each event.")
	PrintInfo("After sync: 'shipwright memory mark-synced'")
}

func memoryDisable(args []string) {
	integrations, err := harness.LoadIntegrations()
	if err != nil {
		Fail(fmt.Sprintf("cannot load integrations: %s", err))
	}

	if !integrations.IsEngramEnabled() {
		PrintInfo("Engram already disabled.")
		return
	}

	integrations.DisableEngram()
	if err := integrations.Save(); err != nil {
		Fail(fmt.Sprintf("cannot save integrations: %s", err))
	}

	PrintSuccess("Engram memory disabled.")
	PrintInfo("Future memory events will be written to progress/decisions.md")
}

func memoryFlush(args []string) {
	integrations, err := harness.LoadIntegrations()
	if err != nil {
		Fail(fmt.Sprintf("cannot load integrations: %s", err))
	}

	if !integrations.IsEngramEnabled() {
		Fail("Engram is not enabled. Run 'shipwright memory enable' first.")
	}

	adapter := harness.NewEngramMemoryAdapter()
	pending := adapter.PendingEvents()

	if len(pending) == 0 {
		PrintInfo("No pending events to flush.")
		return
	}

	PrintSuccess(fmt.Sprintf("Flushing %d pending memory events:", len(pending)))
	fmt.Println()

	for _, ev := range pending {
		output, _ := json.MarshalIndent(ev, "", "  ")
		fmt.Println(string(output))
		fmt.Println("---")
	}

	fmt.Println()
	fmt.Printf("To mark all as synced after processing: shipwright memory mark-synced\n")
}

func memoryMarkSynced(args []string) {
	integrations, err := harness.LoadIntegrations()
	if err != nil {
		Fail(fmt.Sprintf("cannot load integrations: %s", err))
	}

	if !integrations.IsEngramEnabled() {
		Fail("Engram is not enabled.")
	}

	adapter := harness.NewEngramMemoryAdapter()
	if err := adapter.MarkAllSynced(); err != nil {
		Fail(fmt.Sprintf("cannot mark synced: %s", err))
	}

	PrintSuccess("All pending memory events marked as synced.")
}
