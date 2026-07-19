package cmd

import (
	"fmt"

	"shipwright/pkg/harness"
)

func Integrations(args []string) {
	EnsureHarness()

	if len(args) == 0 {
		PrintUsage()
		Fail("usage: shipwright integrations <status|enable|disable|detect>")
	}

	subcommand := args[0]
	rest := args[1:]

	switch subcommand {
	case "status":
		integrationsStatus(rest)
	case "enable":
		integrationsEnable(rest)
	case "disable":
		integrationsDisable(rest)
	case "detect":
		integrationsDetect(rest)
	default:
		Fail(fmt.Sprintf("unknown integrations subcommand: %s\n\nValid: status | enable | disable | detect", subcommand))
	}
}

func integrationsStatus(args []string) {
	integrations, err := harness.LoadIntegrations()
	if err != nil {
		Fail(fmt.Sprintf("cannot load integrations: %s", err))
	}

	fmt.Println("Shipwright — Integrations Status")
	fmt.Println("=====================================")
	fmt.Println()

	if integrations.Platform.OS != "" {
		fmt.Println("Platform:")
		fmt.Printf("  os:       %s\n", integrations.Platform.OS)
		fmt.Printf("  arch:     %s\n", integrations.Platform.Arch)
		fmt.Printf("  ci:       %s\n", boolYesNo(integrations.Platform.IsCI))
		fmt.Println()
	}

	fmt.Println("Engram (Memory):")
	fmt.Printf("  enabled:  %s\n", boolEnabled(integrations.Engram.Enabled))
	fmt.Printf("  mode:     %s\n", integrations.Engram.Mode)
	fmt.Printf("  status:   %s\n", integrations.Engram.Status)
	fmt.Printf("  fallback: %s\n", integrations.Engram.Fallback)
	if integrations.Engram.BinaryPath != "" {
		fmt.Printf("  binary:   %s\n", integrations.Engram.BinaryPath)
	}
	if integrations.Engram.Reason != "" {
		fmt.Printf("  reason:   %s\n", integrations.Engram.Reason)
	}
	fmt.Println()

	fmt.Println("OpenPencil (Design):")
	fmt.Printf("  enabled:  %s\n", boolEnabled(integrations.OpenPencil.Enabled))
	fmt.Printf("  mode:     %s\n", integrations.OpenPencil.Mode)
	fmt.Printf("  status:   %s\n", integrations.OpenPencil.Status)
	fmt.Printf("  fallback: %s\n", integrations.OpenPencil.Fallback)
	if integrations.OpenPencil.AppPath != "" {
		fmt.Printf("  app:      %s\n", integrations.OpenPencil.AppPath)
	}
	if integrations.OpenPencil.MCPServerPath != "" {
		fmt.Printf("  mcp:      %s\n", integrations.OpenPencil.MCPServerPath)
	}
	if integrations.OpenPencil.MCPCommand != "" {
		fmt.Printf("  command:  %s\n", integrations.OpenPencil.MCPCommand)
	}
	if integrations.OpenPencil.Reason != "" {
		fmt.Printf("  reason:   %s\n", integrations.OpenPencil.Reason)
	}
	fmt.Println()

	if integrations.IsEngramEnabled() {
		adapter := harness.NewEngramMemoryAdapter()
		total, pending, synced := adapter.Stats()
		fmt.Printf("Engram queue: %d total, %d pending, %d synced\n", total, pending, synced)
	} else {
		localCount := harness.CountLocalEntries()
		fmt.Printf("Engram local entries: %d in %s\n", localCount, harness.DecisionsFile)
	}
	fmt.Println()

	mode, fallbackUsed, _ := harness.LoadDesignState()
	if mode != "" {
		fmt.Printf("Design mode: %s", mode)
		if fallbackUsed {
			fmt.Print(" (fallback was used)")
		}
		fmt.Println()
	} else {
		fmt.Println("Design mode: (not started)")
	}
}

func integrationsEnable(args []string) {
	if len(args) == 0 {
		Fail("usage: shipwright integrations enable <engram|openpencil>")
	}

	target := args[0]
	integrations, err := harness.LoadIntegrations()
	if err != nil {
		Fail(fmt.Sprintf("cannot load integrations: %s", err))
	}

	switch target {
	case "engram":
		if integrations.IsEngramEnabled() {
			PrintInfo("Engram already enabled.")
			return
		}
		integrations.EnableEngram()
		if err := integrations.Save(); err != nil {
			Fail(fmt.Sprintf("cannot save: %s", err))
		}
		PrintSuccess("Engram enabled. Events will queue in .harness/memory-queue.json")

	case "openpencil":
		if integrations.IsOpenPencilEnabled() {
			PrintInfo("OpenPencil already enabled.")
			return
		}
		integrations.EnableOpenPencil()
		if err := integrations.Save(); err != nil {
			Fail(fmt.Sprintf("cannot save: %s", err))
		}
		PrintSuccess("OpenPencil enabled. Design will use pencil_* MCP tools when available.")
		PrintInfo("Run 'shipwright integrations detect' to check if OpenPencil is installed.")

	default:
		Fail(fmt.Sprintf("unknown integration: %s\n\nValid: engram | openpencil", target))
	}
}

func integrationsDisable(args []string) {
	if len(args) == 0 {
		Fail("usage: shipwright integrations disable <engram|openpencil>")
	}

	target := args[0]
	integrations, err := harness.LoadIntegrations()
	if err != nil {
		Fail(fmt.Sprintf("cannot load integrations: %s", err))
	}

	switch target {
	case "engram":
		if !integrations.IsEngramEnabled() {
			PrintInfo("Engram already disabled.")
			return
		}
		integrations.DisableEngram()
		if err := integrations.Save(); err != nil {
			Fail(fmt.Sprintf("cannot save: %s", err))
		}
		PrintSuccess("Engram disabled. Events will write to progress/decisions.md")

	case "openpencil":
		if !integrations.IsOpenPencilEnabled() {
			PrintInfo("OpenPencil already disabled.")
			return
		}
		integrations.DisableOpenPencil()
		if err := integrations.Save(); err != nil {
			Fail(fmt.Sprintf("cannot save: %s", err))
		}
		PrintSuccess("OpenPencil disabled. Design will use doc-only mode.")

	default:
		Fail(fmt.Sprintf("unknown integration: %s\n\nValid: engram | openpencil", target))
	}
}

func integrationsDetect(args []string) {
	integrations, err := harness.LoadIntegrations()
	if err != nil {
		Fail(fmt.Sprintf("cannot load integrations: %s", err))
	}

	fmt.Println("Shipwright — Integration Detection")
	fmt.Println("======================================")
	fmt.Println()

	probe := harness.RealSystemProbe{}
	config, err := harness.LoadEffectivePortableConfig(probe)
	if err != nil {
		Fail(fmt.Sprintf("cannot load portable config: %s", err))
	}
	integrations.ApplyPortableConfig(config)
	engram := harness.DetectEngramWithConfig(probe, config)
	openpencil := harness.DetectOpenPencilWithConfig(probe, config)
	integrations.ApplyDetection(engram, openpencil)

	engramInstalled := engram.Installed
	fmt.Println("Engram:")
	fmt.Printf("  MCP available: %s\n", boolYesNo(engramInstalled))
	fmt.Printf("  status:        %s\n", engram.Status)
	if engram.Path != "" {
		fmt.Printf("  path:          %s\n", engram.Path)
	}
	if engram.Reason != "" {
		fmt.Printf("  reason:        %s\n", engram.Reason)
	}
	if engramInstalled && !integrations.IsEngramEnabled() {
		fmt.Printf("  → recommendation: enable with 'shipwright integrations enable engram'\n")
	} else if !engramInstalled {
		fmt.Printf("  → fallback: %s\n", engram.Fallback)
	}
	fmt.Println()

	opInstalled := openpencil.Installed
	fmt.Println("OpenPencil:")
	fmt.Printf("  App installed:  %s\n", boolYesNo(opInstalled))
	fmt.Printf("  status:         %s\n", openpencil.Status)
	if openpencil.Path != "" {
		fmt.Printf("  mcp path:       %s\n", openpencil.Path)
	}
	if openpencil.Reason != "" {
		fmt.Printf("  reason:         %s\n", openpencil.Reason)
	}
	if opInstalled && openpencil.Active {
		fmt.Printf("  Canvas active:   yes\n")
		fmt.Printf("  → recommendation: enable with 'shipwright integrations enable openpencil'\n")
	} else if opInstalled {
		fmt.Printf("  Canvas active:   no\n")
		fmt.Printf("  → fallback: %s (canvas not active)\n", openpencil.Fallback)
	} else {
		fmt.Printf("  → fallback: %s\n", openpencil.Fallback)
	}

	if err := integrations.Save(); err != nil {
		PrintInfo(fmt.Sprintf("warning: could not save detection results: %s", err))
	}
}

func boolYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
