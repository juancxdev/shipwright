package cmd

import (
	"fmt"
	"strings"

	"shipwright/pkg/harness"
)

func Next(args []string) {
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

	result := harness.Advance(state)

	if !result.Transitioned {
		if result.BlockReason != "" {
			fmt.Println("BLOCKED")
			fmt.Println()
			fmt.Println(result.BlockReason)
			if len(result.MissingArtifacts) > 0 {
				fmt.Println()
				fmt.Println("Artefactos faltantes:")
				for _, a := range result.MissingArtifacts {
					fmt.Printf("  ✗ %s\n", a)
				}
			}
		} else {
			fmt.Println(result.Message)
		}

		state.SetBlocked(result.BlockReason)
		if result.BlockReason == "" {
			state.SetReady()
		}
		_ = state.Save()
		_ = harness.UpdateCurrent(state, computeNextAction(state))
		Exit(1)
		return
	}

	if err := state.Save(); err != nil {
		Fail(fmt.Sprintf("error guardando estado: %s", err))
	}

	nextAction := computeNextAction(state)
	if err := harness.UpdateCurrent(state, nextAction); err != nil {
		Fail(fmt.Sprintf("error actualizando progress: %s", err))
	}

	details := result.Message
	if len(result.MissingArtifacts) > 0 {
		details += " (faltaban: " + strings.Join(result.MissingArtifacts, ", ") + ")"
	}
	if err := harness.AppendHistory("next", result.To, details); err != nil {
		Fail(fmt.Sprintf("error registrando history: %s", err))
	}

	integrations, _ := harness.LoadIntegrations()
	memService := harness.NewMemoryService(integrations)
	if err := harness.SavePhaseTransitionMemory(memService, state, result.From, result.To); err != nil {
		PrintInfo(fmt.Sprintf("warning: no se pudo guardar memoria: %s", err))
	}

	fromAgent := harness.ActiveAgentForPhase(result.From)
	toAgent := harness.ActiveAgentForPhase(result.To)
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
			Phase:     result.To,
			Action:    "phase transition",
			Reason:    result.Message,
		})
	}

	PrintSuccess(fmt.Sprintf("Fase: %s → %s", result.From, result.To))
	PrintInfo(result.Message)
	fmt.Println()
	fmt.Println("Next action:")
	fmt.Println(nextAction)
}
