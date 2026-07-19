package cmd

import (
	"fmt"

	"shipwright/pkg/harness"
)

func Design(args []string) {
	EnsureHarness()

	if len(args) == 0 {
		PrintUsage()
		Fail("usage: shipwright design <start|status>")
	}

	subcommand := args[0]
	rest := args[1:]

	switch subcommand {
	case "start":
		designStart(rest)
	case "status":
		designStatus(rest)
	default:
		Fail(fmt.Sprintf("unknown design subcommand: %s\n\nValid: start | status", subcommand))
	}
}

func designStart(args []string) {
	state, err := harness.LoadState()
	if err != nil {
		Fail(err.Error())
	}

	if state.CurrentPhase != harness.StateUXDecision &&
		state.CurrentPhase != harness.StateUXDesign {
		Fail(fmt.Sprintf("design start solo es válido en UX_DECISION o UX_DESIGN. Fase actual: %s", state.CurrentPhase))
	}

	if state.RequiresUI == nil {
		Fail("requires_ui no está decidido. Seteá requires_ui en .harness/state.json a true o false primero.")
	}

	if !*state.RequiresUI {
		Fail("requires_ui es false. Este proyecto no requiere diseño UI/UX.")
	}

	integrations, _ := harness.LoadIntegrations()
	designService := harness.NewDesignService(integrations)

	request := state.InitialRequest
	if request == "" {
		request = state.ProjectName
	}

	result, err := designService.StartDesign(state, request)
	if err != nil {
		Fail(fmt.Sprintf("design start failed: %s", err))
	}

	if err := harness.AppendHistory("design:start", state.CurrentPhase,
		fmt.Sprintf("Design started via %s. %s", result.Adapter, result.Message)); err != nil {
		PrintInfo(fmt.Sprintf("warning: could not log to history: %s", err))
	}

	integrations2, _ := harness.LoadIntegrations()
	memService := harness.NewMemoryService(integrations2)
	if result.FallbackUsed {
		_ = memService.SaveDiscovery(
			"Design started in doc-only mode: "+state.ProjectName,
			"design/ux-approval",
			"Design phase started with doc-only fallback (OpenPencil unavailable)",
			"OpenPencil not available — generated text-based wireframes and prototype",
			"design/wireframes.md, design/prototype.md, design/responsive-qa.md",
			"OpenPencil can be enabled later with 'shipwright integrations enable openpencil'",
		)
	} else {
		_ = memService.SaveDecision(
			"Design started with OpenPencil: "+state.ProjectName,
			"design/ux-approval",
			"Design phase started with OpenPencil adapter",
			"OpenPencil available — design task created for AI agent",
			"design/openpencil/design-task.md, design/openpencil/app.pen",
			"",
		)
	}

	PrintSuccess(fmt.Sprintf("Design started via %s adapter", result.Adapter))
	PrintInfo(result.Message)
	fmt.Println()
	fmt.Println("Files created:")
	for _, f := range result.FilesCreated {
		fmt.Printf("  ✓ %s\n", f)
	}

	if result.TaskFile != "" {
		fmt.Println()
		fmt.Println("Next steps for AI agent:")
		fmt.Printf("  1. Read %s\n", result.TaskFile)
		fmt.Println("  2. Use open-pencil_* MCP tools to create responsive mobile/tablet/desktop frames")
		fmt.Println("  3. Export wireframes to design/openpencil/exports/")
		fmt.Println("  4. Inspect screenshots for overflow/clipping and create design/responsive-qa.md")
		fmt.Println("  5. Create design/prototype.md describing the visual design")
		fmt.Println("  6. Run: shipwright design status")
	} else {
		fmt.Println()
		fmt.Println("Doc-only mode:")
		fmt.Println("  Edit the generated documents to describe the UX design.")
		fmt.Println("  Then run: shipwright next")
	}

	fmt.Println()
	fmt.Println("After design is complete:")
	fmt.Println("  shipwright next       (advance to UX_APPROVAL)")
	fmt.Println("  shipwright approve ux-design  (user must approve)")
}

func designStatus(args []string) {
	integrations, _ := harness.LoadIntegrations()
	designService := harness.NewDesignService(integrations)

	status, err := designService.Status()
	if err != nil {
		Fail(fmt.Sprintf("cannot get design status: %s", err))
	}

	fmt.Println("Shipwright — Design Status")
	fmt.Println("==============================")
	fmt.Println()

	fmt.Printf("Adapter:       %s\n", status.Adapter)
	fmt.Printf("Mode:          %s\n", status.Mode)
	fmt.Printf("Available:     %s\n", boolYesNo(status.Available))
	if status.PenFile != "" {
		penExists := harness.ArtifactExists(status.PenFile)
		fmt.Printf("Pen file:      %s (exists: %s)\n", status.PenFile, boolYesNo(penExists))
	}
	fmt.Println()

	fmt.Println("Artifacts:")
	printArtifactStatus("design/ux-brief.md", status.HasBrief)
	printArtifactStatus("design/user-flows.md", status.HasFlows)
	printArtifactStatus("design/design-decisions.md", status.HasDecisions)
	printArtifactStatus("design/wireframes.md", status.HasWireframes)
	printArtifactStatus("design/prototype.md", status.HasPrototype)
	printArtifactStatus("design/responsive-qa.md", status.HasResponsiveQA)

	if status.HasTaskFile {
		printArtifactStatus(harness.DesignTaskFile, true)
	}

	fmt.Println()

	state, _ := harness.LoadState()
	if state != nil {
		if state.CurrentPhase == harness.StateUXDesign || state.CurrentPhase == harness.StateUXApproval {
			allReady := status.HasBrief && status.HasFlows && status.HasPrototype && status.HasResponsiveQA
			if allReady {
				fmt.Println("✓ All required design artifacts present.")
				if state.CurrentPhase == harness.StateUXDesign {
					fmt.Println("  Run: shipwright next  (advance to UX_APPROVAL)")
				}
				if state.CurrentPhase == harness.StateUXApproval {
					fmt.Println("  Run: shipwright approve ux-design  (user approval)")
				}
			} else {
				fmt.Println("✗ Missing required artifacts for UX_APPROVAL.")
				fmt.Println("  Need: design/ux-brief.md, design/user-flows.md, design/prototype.md, design/responsive-qa.md")
			}
		}
	}
}

func printArtifactStatus(path string, exists bool) {
	mark := "✗"
	if exists {
		mark = "✓"
	}
	fmt.Printf("  %s %s\n", mark, path)
}
