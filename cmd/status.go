package cmd

import (
	"fmt"
	"strings"

	"shipwright/pkg/harness"
)

func Status(args []string) {
	EnsureHarness()

	state, err := harness.LoadState()
	if err != nil {
		Fail(err.Error())
	}

	fmt.Println("Shipwright — Status")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println()
	fmt.Printf("Project:   %s\n", state.ProjectName)
	fmt.Printf("Phase:     %s\n", state.CurrentPhase)
	fmt.Printf("Status:    %s\n", state.Status)

	activeAgent := harness.ActiveAgentForPhase(state.CurrentPhase)
	if activeAgent != nil {
		fmt.Printf("Agent:     %s\n", activeAgent.Name)
	}

	if state.InitialRequest != "" {
		fmt.Printf("Request:   %s\n", truncate(state.InitialRequest, 60))
	}
	fmt.Println()

	fmt.Println("Approvals:")
	gates := []struct {
		key   string
		label string
	}{
		{harness.GateScope, "scope"},
		{harness.GateUXDesign, "ux-design"},
		{harness.GateTechnicalPlan, "technical-plan"},
		{harness.GateTechLeadReview, "tech-lead"},
		{harness.GateFinalAcceptance, "final-acceptance"},
	}
	for _, g := range gates {
		mark := "[ ]"
		if state.IsApproved(g.key) {
			mark = "[x]"
		}
		fmt.Printf("  %s %s\n", mark, g.label)
	}
	fmt.Println()

	if state.RequiresUI != nil {
		ui := "no"
		if *state.RequiresUI {
			ui = "yes"
		}
		fmt.Printf("Requires UI: %s\n", ui)
	}

	if state.ActiveChangeRequest != nil && *state.ActiveChangeRequest != "" {
		fmt.Printf("Active CR:   %s\n", *state.ActiveChangeRequest)
	}

	integrations, _ := harness.LoadIntegrations()
	if integrations != nil {
		fmt.Println()
		fmt.Println("Integrations:")
		fmt.Printf("  engram:      %s (%s)\n", boolEnabled(integrations.Engram.Enabled), integrations.Engram.Status)
		fmt.Printf("  openpencil:  %s (%s)\n", boolEnabled(integrations.OpenPencil.Enabled), integrations.OpenPencil.Status)

		fmt.Println()
		fmt.Println("Memory:")
		memMode := "fallback"
		if integrations.IsEngramEnabled() {
			memMode = "engram"
		}
		fmt.Printf("  adapter:     %s\n", memMode)

		if integrations.IsEngramEnabled() {
			adapter := harness.NewEngramMemoryAdapter()
			total, pending, synced := adapter.Stats()
			fmt.Printf("  queue:       %d total, %d pending, %d synced\n", total, pending, synced)
		} else {
			localCount := harness.CountLocalEntries()
			fmt.Printf("  local log:   %s (%d entries)\n", harness.DecisionsFile, localCount)
		}

		fmt.Println()
		fmt.Println("Design:")
		designMode := "doc-only"
		if integrations.IsOpenPencilEnabled() {
			designMode = "openpencil"
		}
		fmt.Printf("  adapter:     %s\n", designMode)
		fmt.Printf("  openpencil:  %s (%s)\n", boolEnabled(integrations.OpenPencil.Enabled), integrations.OpenPencil.Status)

		dStatus, _ := harness.NewDesignService(integrations).Status()
		if dStatus != nil {
			fmt.Printf("  brief:       %s\n", boolCheck(dStatus.HasBrief))
			fmt.Printf("  user-flows:  %s\n", boolCheck(dStatus.HasFlows))
			fmt.Printf("  decisions:   %s\n", boolCheck(dStatus.HasDecisions))
			fmt.Printf("  prototype:   %s\n", boolCheck(dStatus.HasPrototype))
			fmt.Printf("  responsive:  %s\n", boolCheck(dStatus.HasResponsiveQA))
			if dStatus.PenFile != "" {
				penExists := harness.ArtifactExists(dStatus.PenFile)
				fmt.Printf("  pen file:    %s (%s)\n", dStatus.PenFile, boolCheck(penExists))
			}
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("Next action:")

	action := computeNextAction(state)
	fmt.Println(action)
}

func boolEnabled(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

func boolCheck(b bool) string {
	if b {
		return "✓"
	}
	return "✗"
}

func computeNextAction(state *harness.State) string {
	switch state.CurrentPhase {
	case harness.StateIntake:
		return "Run: shipwright start \"<your request>\""

	case harness.StateDiscovery:
		missing := harness.CheckArtifacts([]string{"product/context.md", "product/assumptions.md", "product/open-questions.md"})
		if len(missing) > 0 {
			return discoveryNextAction(missing)
		}
		return "Product discovery artifacts exist. If critical questions are resolved, run: shipwright next"

	case harness.StateProductContextReady:
		missing := harness.CheckArtifacts([]string{"architecture/technology-options.md"})
		if len(missing) > 0 {
			return fmt.Sprintf("Blocked — missing:\n  %s\n\nRun: shipwright scaffold", strings.Join(missing, "\n  "))
		}
		return "Gates met. Run: shipwright next  (or: shipwright run)"

	case harness.StateTechnicalScopeDraft:
		missing := harness.CheckArtifacts([]string{"product/scope.md"})
		if len(missing) > 0 {
			return fmt.Sprintf("Blocked — missing:\n  %s\n\nRun: shipwright scaffold", strings.Join(missing, "\n  "))
		}
		return "Gates met. Run: shipwright next  (or: shipwright run)"

	case harness.StateScopeReview:
		if !state.IsApproved(harness.GateScope) {
			return "Approval gate — present product/scope.md to the user and ask: approve scope or request changes. If user approves, run: shipwright approve scope. If user wants changes, run: shipwright request-change \"<reason>\""
		}
		return "Scope approved. Run: shipwright next"

	case harness.StateScopeApproved:
		missing := harness.CheckArtifacts([]string{"project/project-charter.md", "project/project-plan.md", "project/risk-register.md"})
		if len(missing) > 0 {
			return fmt.Sprintf("Blocked — missing:\n  %s\n\nRun: shipwright scaffold", strings.Join(missing, "\n  "))
		}
		return "Gates met. Run: shipwright next  (or: shipwright run)"

	case harness.StateProjectPlanning:
		missing := harness.CheckArtifacts([]string{"project/delivery-plan.md"})
		if len(missing) > 0 {
			return fmt.Sprintf("Blocked — missing:\n  %s\n\nRun: shipwright scaffold", strings.Join(missing, "\n  "))
		}
		return "Gates met. Run: shipwright next  (or: shipwright run)"

	case harness.StateUXDecision:
		if state.RequiresUI == nil {
			return "Blocked — set requires_ui in .harness/state.json to true or false"
		}
		if *state.RequiresUI {
			missing := harness.CheckArtifacts([]string{"design/ux-brief.md"})
			if len(missing) > 0 {
				return "Blocked — run: shipwright design start\n  (generates ux-brief.md, user-flows.md, design-decisions.md)\n\nOr: shipwright run  (auto-runs design start)"
			}
		}
		return "Gates met. Run: shipwright next  (or: shipwright run)"

	case harness.StateUXDesign:
		missing := harness.CheckArtifacts([]string{"design/prototype.md", "design/user-flows.md", "design/responsive-qa.md"})
		if len(missing) > 0 {
			return fmt.Sprintf("Blocked — missing:\n  %s\n\nRun: shipwright scaffold", strings.Join(missing, "\n  "))
		}
		return "Gates met. Run: shipwright next  (or: shipwright run)"

	case harness.StateUXApproval:
		if !state.IsApproved(harness.GateUXDesign) {
			return "Approval gate.\nTo approve: shipwright approve ux-design\nTo request changes: shipwright request-change \"<reason>\""
		}
		return "Run: shipwright next"

	case harness.StateTechnicalDesign:
		missing := harness.CheckArtifacts([]string{
			"architecture/system-architecture.md",
			"contracts/openapi.yaml",
			"backlog/epics.md",
			"backlog/user-stories.md",
			"backlog/frontend-tasks.md",
			"backlog/backend-tasks.md",
			"sdd/proposal.md",
			"sdd/spec.md",
			"sdd/tasks.md",
		})
		if len(missing) > 0 {
			if harness.ArtifactExists("contracts/openapi.yaml") &&
				(!harness.ArtifactExists("backlog/frontend-tasks.md") || !harness.ArtifactExists("backlog/backend-tasks.md")) {
				return fmt.Sprintf("Blocked — missing:\n  %s\n\nRun: shipwright contract validate\nThen: shipwright contract generate-tasks", strings.Join(missing, "\n  "))
			}
			return fmt.Sprintf("Blocked — missing:\n  %s\n\nRun: shipwright scaffold\nIf contract exists, run: shipwright contract generate-tasks", strings.Join(missing, "\n  "))
		}
		if result := harness.ValidateContract(harness.ContractFile); !result.IsValid {
			return "Blocked — invalid API contract.\nRun: shipwright contract validate"
		}
		return "Gates met. Run: shipwright next  (or: shipwright run)"

	case harness.StateBacklogReady:
		if !state.IsApproved(harness.GateTechnicalPlan) {
			return "Approval gate.\nTo approve: shipwright approve technical-plan"
		}
		return "Run: shipwright next"

	case harness.StateImplementation:
		missing := harness.CheckArtifacts([]string{"progress/frontend.md", "progress/backend.md"})
		if len(missing) > 0 {
			return fmt.Sprintf("Blocked — missing:\n  %s\n\nRun: shipwright scaffold (placeholders)", strings.Join(missing, "\n  "))
		}
		if harness.TDDBlockReason() != "" {
			return "Strict TDD evidence is not ready.\nRun: shipwright tdd status\nThen add executed test evidence to progress/frontend.md, progress/backend.md, or reports/tdd-report.md"
		}
		return "Verify contract-first work:\n  shipwright contract check-mocks\n  shipwright contract check-compliance\n  shipwright tdd status\n\nIf all gates pass, run: shipwright next  (or: shipwright run)"

	case harness.StateIntegration:
		missing := harness.CheckArtifacts([]string{"reports/contract-test-report.md", "reports/review-checklist.md"})
		if len(missing) > 0 {
			return fmt.Sprintf("Blocked — missing:\n  %s\n\nRun: shipwright review start", strings.Join(missing, "\n  "))
		}
		if reason := harness.ContractReviewBlockReason(); reason != "" {
			return "Review evidence is not ready.\nRun: shipwright review status"
		}
		return "Gates met. Run: shipwright next  (or: shipwright run)"

	case harness.StateQASecurityReview:
		missing := harness.CheckArtifacts(harness.RequiredReviewArtifacts())
		if len(missing) > 0 {
			return fmt.Sprintf("Blocked — missing:\n  %s\n\nRun: shipwright review start", strings.Join(missing, "\n  "))
		}
		if reason := harness.ReviewBlockReason(); reason != "" {
			return "Review gate blocked.\nRun: shipwright review status\nCritical findings block; medium findings require explicit decision."
		}
		return "Gates met. Run: shipwright next  (or: shipwright run)"

	case harness.StateTechLeadReview:
		if !state.IsApproved(harness.GateTechLeadReview) {
			return "Approval gate.\nTo approve: shipwright approve tech-lead\nTo reject: shipwright request-change \"<reason>\""
		}
		return "Run: shipwright next"

	case harness.StateUserAcceptance:
		if !state.IsApproved(harness.GateFinalAcceptance) {
			missing := harness.CheckArtifacts([]string{"project/acceptance-report.md"})
			if len(missing) > 0 {
				return fmt.Sprintf("Blocked — missing:\n  %s\n\nRun: shipwright scaffold", strings.Join(missing, "\n  "))
			}
			return "Approval gate.\nTo accept: shipwright approve final-acceptance\nTo request changes: shipwright request-change \"<reason>\""
		}
		return "Run: shipwright next"

	case harness.StateClosed:
		return "Project closed. ✅"

	case harness.StateChangeRequest:
		missing := harness.CheckArtifacts([]string{"project/change-management.md"})
		if len(missing) > 0 {
			return fmt.Sprintf("Blocked — missing:\n  %s\n\nRun: shipwright scaffold", strings.Join(missing, "\n  "))
		}
		return "Change request assessed. Run: shipwright next  (or: shipwright run)"

	default:
		return fmt.Sprintf("Unknown phase: %s", state.CurrentPhase)
	}
}
