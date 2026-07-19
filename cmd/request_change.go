package cmd

import (
	"fmt"
	"strings"

	"shipwright/pkg/harness"
)

func RequestChange(args []string) {
	EnsureHarness()

	state, err := harness.LoadState()
	if err != nil {
		Fail(err.Error())
	}

	if state.CurrentPhase == harness.StateIntake {
		Fail("estás en INTAKE. No hay nada que cambiar todavía.")
	}

	if state.CurrentPhase == harness.StateClosed {
		Fail("el proyecto está cerrado. Iniciá un nuevo proyecto con 'shipwright init'.")
	}

	reason := "Cambio solicitado por el usuario"
	if len(args) > 0 {
		reason = strings.Join(args, " ")
	}

	result := harness.RequestChange(state, reason)

	if result.Error != "" {
		PrintError(result.Error)
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

	if err := harness.AppendHistory("request-change", result.To,
		fmt.Sprintf("Change request: %s. %s", truncate(reason, 60), result.Message)); err != nil {
		Fail(fmt.Sprintf("error registrando history: %s", err))
	}

	integrations, _ := harness.LoadIntegrations()
	memService := harness.NewMemoryService(integrations)
	if err := harness.SaveChangeRequestMemory(memService, state, reason, result.CRFile); err != nil {
		PrintInfo(fmt.Sprintf("warning: no se pudo guardar memoria: %s", err))
	} else {
		PrintInfo(fmt.Sprintf("memory saved via %s", memService.AdapterName()))
	}

	PrintSuccess(fmt.Sprintf("Change request creado. Fase: %s → %s", result.From, result.To))
	PrintInfo(fmt.Sprintf("CR file: %s", result.CRFile))
	PrintInfo(result.Message)
	fmt.Println()
	fmt.Println("Next action:")
	fmt.Println(nextAction)
}
