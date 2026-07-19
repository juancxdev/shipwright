package cmd

import (
	"fmt"

	"shipwright/pkg/harness"
)

func Approve(args []string) {
	EnsureHarness()

	if len(args) == 0 {
		Fail("usage: shipwright approve <gate>\n\nGates: scope | ux-design | technical-plan | tech-lead | final-acceptance")
	}

	gate := args[0]
	validGates := map[string]bool{
		harness.GateScope:           true,
		harness.GateUXDesign:        true,
		harness.GateTechnicalPlan:   true,
		harness.GateTechLeadReview:  true,
		harness.GateFinalAcceptance: true,
	}
	if !validGates[gate] {
		Fail(fmt.Sprintf("gate inválido: '%s'\n\nGates válidos: scope | ux-design | technical-plan | tech-lead | final-acceptance", gate))
	}

	state, err := harness.LoadState()
	if err != nil {
		Fail(err.Error())
	}

	result := harness.ApproveGate(state, gate)

	if result.Error != "" {
		if len(result.MissingArtifacts) > 0 {
			fmt.Println("BLOCKED")
			fmt.Println()
			fmt.Println(result.Error)
			fmt.Println()
			fmt.Println("Artefactos faltantes:")
			for _, a := range result.MissingArtifacts {
				fmt.Printf("  ✗ %s\n", a)
			}
		} else {
			PrintError(result.Error)
		}
		Exit(1)
		return
	}

	if err := writeApprovalRecord(state, gate); err != nil {
		Fail(fmt.Sprintf("error guardando approval: %s", err))
	}

	if err := state.Save(); err != nil {
		Fail(fmt.Sprintf("error guardando estado: %s", err))
	}

	nextAction := computeNextAction(state)
	if err := harness.UpdateCurrent(state, nextAction); err != nil {
		Fail(fmt.Sprintf("error actualizando progress: %s", err))
	}

	if err := harness.AppendHistory("approve:"+gate, result.To,
		fmt.Sprintf("Gate '%s' aprobado. %s", gate, result.Message)); err != nil {
		Fail(fmt.Sprintf("error registrando history: %s", err))
	}

	integrations, _ := harness.LoadIntegrations()
	memService := harness.NewMemoryService(integrations)
	if err := harness.SaveGateMemory(memService, state, gate); err != nil {
		PrintInfo(fmt.Sprintf("warning: no se pudo guardar memoria: %s", err))
	} else {
		PrintInfo(fmt.Sprintf("memory saved via %s", memService.AdapterName()))
	}

	fromAgent := harness.ActiveAgentForPhase(result.From)
	toAgent := harness.ActiveAgentForPhase(result.To)
	if fromAgent != nil || toAgent != nil {
		fromName := "user"
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
			Phase:     result.To,
			Action:    "gate approved: " + gate,
			Reason:    result.Message,
		})
	}

	PrintSuccess(fmt.Sprintf("Gate '%s' aprobado. Fase: %s → %s", gate, result.From, result.To))
	PrintInfo(result.Message)
	fmt.Println()
	fmt.Println("Next action:")
	fmt.Println(nextAction)
}

func writeApprovalRecord(state *harness.State, gate string) error {
	content := fmt.Sprintf(`{
  "approval_id": "%s-%s",
  "phase": "%s",
  "approved_by": "user",
  "approved_at": "%s",
  "gate": "%s",
  "notes": "Approved via harness CLI"
}
`, gate, state.ProjectID, state.CurrentPhase, nowTimestamp(), gate)

	path := fmt.Sprintf(".harness/approvals/%s.json", gate)
	return harness.WriteFile(path, content)
}

func nowTimestamp() string {
	return harness.NowISO()
}
