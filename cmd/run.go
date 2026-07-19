package cmd

import (
	"fmt"
	"strings"

	"shipwright/pkg/harness"
)

func Run(args []string) {
	EnsureHarness()

	state, err := harness.LoadState()
	if err != nil {
		Fail(err.Error())
	}

	if state.CurrentPhase == harness.StateIntake {
		Fail("estás en INTAKE. Ejecutá 'shipwright start \"<request>\"' primero.")
	}

	if state.CurrentPhase == harness.StateClosed {
		Fail("el proyecto está cerrado.")
	}

	fmt.Println("Shipwright — Vertical Slice Run")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	steps := 0
	maxSteps := 30

	for {
		if steps >= maxSteps {
			fmt.Println("Reached max steps (30). Stopping.")
			break
		}
		steps++

		fmt.Printf("[Step %d] Phase: %s\n", steps, state.CurrentPhase)

		if state.CurrentPhase == harness.StateClosed {
			fmt.Println()
			PrintSuccess("Project closed! Vertical slice complete.")
			break
		}

		if state.CurrentPhase == harness.StateUXDecision && state.RequiresUI == nil {
			fmt.Println()
			fmt.Println("  ⏸  Blocked: requires_ui not decided")
			fmt.Println("  Set requires_ui in .harness/state.json to true or false")
			fmt.Println("  Then run: shipwright run")
			_ = state.Save()
			break
		}

		fmt.Printf("  → Scaffolding artifacts for %s...\n", state.CurrentPhase)
		scaffoldResult := harness.ScaffoldPhase(state)
		if len(scaffoldResult.Generated) > 0 {
			for _, f := range scaffoldResult.Generated {
				fmt.Printf("    ✓ %s\n", f)
			}
		}
		if len(scaffoldResult.Skipped) > 0 {
			for _, f := range scaffoldResult.Skipped {
				fmt.Printf("    → %s (exists)\n", f)
			}
		}
		if len(scaffoldResult.Errors) > 0 {
			for _, e := range scaffoldResult.Errors {
				fmt.Printf("    ✗ %s\n", e)
			}
			fmt.Println("  Stopping due to scaffold errors.")
			break
		}

		if state.CurrentPhase == harness.StateUXDecision && state.RequiresUI != nil && *state.RequiresUI {
			fmt.Println("  → Running design start (UI required)...")
			integrations, _ := harness.LoadIntegrations()
			designService := harness.NewDesignService(integrations)
			req := state.InitialRequest
			if req == "" {
				req = state.ProjectName
			}
			designResult, err := designService.StartDesign(state, req)
			if err != nil {
				fmt.Printf("    ✗ design start failed: %s\n", err)
				break
			}
			fmt.Printf("    ✓ design started via %s\n", designResult.Adapter)
		}

		if isApprovalPhase(state.CurrentPhase) {
			gate := gateForPhase(state.CurrentPhase)
			fmt.Println()
			fmt.Printf("  ⏸  Approval gate: %s\n", gate)
			fmt.Println("  Run one of:")
			fmt.Printf("    shipwright approve %s\n", gate)

			if state.CurrentPhase == harness.StateScopeReview || state.CurrentPhase == harness.StateUXApproval {
				fmt.Println("    shipwright request-change \"<reason>\"")
			}
			if state.CurrentPhase == harness.StateTechLeadReview {
				fmt.Println("    shipwright request-change \"<reason>\"")
			}
			if state.CurrentPhase == harness.StateUserAcceptance {
				fmt.Println("    shipwright request-change \"<reason>\"")
			}

			_ = state.Save()
			_ = harness.UpdateCurrent(state, "")
			break
		}

		fmt.Printf("  → Advancing...\n")
		advanceResult := harness.Advance(state)

		if !advanceResult.Transitioned {
			if advanceResult.BlockReason != "" {
				fmt.Println()
				fmt.Println("  ⏸  BLOCKED")
				fmt.Printf("  %s\n", advanceResult.BlockReason)
				if len(advanceResult.MissingArtifacts) > 0 {
					fmt.Println()
					fmt.Println("  Missing artifacts:")
					for _, a := range advanceResult.MissingArtifacts {
						fmt.Printf("    ✗ %s\n", a)
					}
					fmt.Println()
					fmt.Println("  Run: shipwright scaffold")
					fmt.Println("  Then: shipwright run")
				}
			} else {
				fmt.Printf("  %s\n", advanceResult.Message)
			}

			if advanceResult.BlockReason == "" {
				state.SetReady()
			}
			_ = state.Save()
			_ = harness.UpdateCurrent(state, "")
			break
		}

		if err := state.Save(); err != nil {
			Fail(fmt.Sprintf("error saving state: %s", err))
		}

		nextAction := computeNextAction(state)
		_ = harness.UpdateCurrent(state, nextAction)

		details := advanceResult.Message
		_ = harness.AppendHistory("run:next", advanceResult.To, details)

		integrations, _ := harness.LoadIntegrations()
		memService := harness.NewMemoryService(integrations)
		_ = harness.SavePhaseTransitionMemory(memService, state, advanceResult.From, advanceResult.To)

		fromAgent := harness.ActiveAgentForPhase(advanceResult.From)
		toAgent := harness.ActiveAgentForPhase(advanceResult.To)
		if fromAgent != nil || toAgent != nil {
			fromName := "none"
			if fromAgent != nil {
				fromName = fromAgent.Name
			}
			toName := "none"
			if toAgent != nil {
				toName = toAgent.Name
			}
			_ = harness.LogHandoff(harness.HandoffRecord{
				FromAgent: fromName,
				ToAgent:   toName,
				Phase:     advanceResult.To,
				Action:    "phase transition",
				Reason:    advanceResult.Message,
			})
		}

		fmt.Printf("    ✓ %s → %s\n", advanceResult.From, advanceResult.To)
		fmt.Printf("    %s\n", advanceResult.Message)
		fmt.Println()
	}
}

func isApprovalPhase(phase string) bool {
	return phase == harness.StateScopeReview ||
		phase == harness.StateUXApproval ||
		phase == harness.StateBacklogReady ||
		phase == harness.StateTechLeadReview ||
		phase == harness.StateUserAcceptance
}

func gateForPhase(phase string) string {
	switch phase {
	case harness.StateScopeReview:
		return harness.GateScope
	case harness.StateUXApproval:
		return harness.GateUXDesign
	case harness.StateBacklogReady:
		return harness.GateTechnicalPlan
	case harness.StateTechLeadReview:
		return harness.GateTechLeadReview
	case harness.StateUserAcceptance:
		return harness.GateFinalAcceptance
	default:
		return "(unknown)"
	}
}
