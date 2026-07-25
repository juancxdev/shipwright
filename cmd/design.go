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
			".harness/artifacts/design/ux-approval",
			"Design phase started with doc-only fallback",
			"Primary design provider unavailable — generated text-based wireframes and prototype",
			".harness/artifacts/design/wireframes.md, .harness/artifacts/design/prototype.md, .harness/artifacts/design/responsive-qa.md",
			"Set STITCH_API_KEY and enable Stitch for high-fidelity design",
		)
	} else {
		_ = memService.SaveDecision(
			"Design started with "+result.Adapter+": "+state.ProjectName,
			".harness/artifacts/design/ux-approval",
			"Design phase started with "+result.Adapter+" adapter",
			"Design task created for AI agent",
			result.TaskFile,
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
		if result.Adapter == harness.DesignModeStitch {
			fmt.Println("  2. Use Google Stitch SDK/MCP to generate responsive mobile/tablet/desktop screens")
			fmt.Println("  3. Export screenshots to .harness/artifacts/design/stitch/exports/")
			fmt.Println("  4. Export HTML to .harness/artifacts/design/stitch/html/ when available")
			fmt.Println("  5. Create prototype.md, responsive-qa.md, stitch-report.md, and code-component-map.md")
		} else if result.Adapter == harness.DesignModeOpenDesign {
			fmt.Println("  2. Use OpenDesign MCP tools as artifact tools: open-design_get_active_context, open-design_list_projects, open-design_create_artifact")
			fmt.Println("  3. Generate .harness/artifacts/design/opendesign/<entry>.html plus <entry>.html.artifact.json")
			fmt.Println("  4. Publish/create the artifact through OpenDesign MCP when available")
			fmt.Println("  5. Create prototype.md, responsive-qa.md, opendesign-report.md, fidelity-report.md, and code-component-map.md when needed")
		} else {
			fmt.Println("  2. Use the selected design provider to create responsive mobile/tablet/desktop evidence")
			fmt.Println("  3. Inspect screenshots for overflow/clipping and create .harness/artifacts/design/responsive-qa.md")
			fmt.Println("  4. Create .harness/artifacts/design/prototype.md describing the visual design")
		}
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
	printArtifactStatus(".harness/artifacts/design/ux-brief.md", status.HasBrief)
	printArtifactStatus(".harness/artifacts/design/user-flows.md", status.HasFlows)
	printArtifactStatus(".harness/artifacts/design/design-decisions.md", status.HasDecisions)
	printArtifactStatus(".harness/artifacts/design/wireframes.md", status.HasWireframes)
	printArtifactStatus(".harness/artifacts/design/prototype.md", status.HasPrototype)
	printArtifactStatus(".harness/artifacts/design/responsive-qa.md", status.HasResponsiveQA)

	if status.HasTaskFile {
		switch status.Adapter {
		case harness.DesignModeStitch:
			printArtifactStatus(harness.DesignStitchTaskFile, true)
		case harness.DesignModeOpenDesign:
			printArtifactStatus(harness.DesignOpenDesignTaskFile, true)
		default:
			printArtifactStatus(harness.DesignTaskFile, true)
		}
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
				fmt.Println("  Need: .harness/artifacts/design/ux-brief.md, .harness/artifacts/design/user-flows.md, .harness/artifacts/design/prototype.md, .harness/artifacts/design/responsive-qa.md")
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
