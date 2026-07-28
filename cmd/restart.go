package cmd

import (
	"fmt"
	"strings"

	"shipwright/internal/app/harness"
)

func Restart(args []string) {
	EnsureHarness()

	if len(args) == 0 {
		Fail("usage: shipwright restart \"<new request>\"")
	}

	request := strings.Join(args, " ")
	if strings.TrimSpace(request) == "" {
		Fail("la petición no puede estar vacía.")
	}

	state, err := harness.LoadState()
	if err != nil {
		Fail(err.Error())
	}

	result, err := harness.RestartRequest(state, request)
	if err != nil {
		Fail(err.Error())
	}

	if err := result.State.Save(); err != nil {
		Fail(fmt.Sprintf("error guardando estado: %s", err))
	}

	nextAction := "Product Owner discovery round required for the restarted delivery cycle. Open OpenCode, run /shipwright-active-agent, let product-owner ask discovery questions in chat, then generate .harness/artifacts/product/context.md, .harness/artifacts/product/assumptions.md, .harness/artifacts/product/open-questions.md and .harness/artifacts/product/scope.md."
	if err := harness.UpdateCurrent(result.State, nextAction); err != nil {
		Fail(fmt.Sprintf("error actualizando progress: %s", err))
	}

	if err := harness.AppendHistory("restart", harness.StateDiscovery,
		fmt.Sprintf("Ciclo reiniciado desde %s. Nueva petición: %s. Backup: %s", result.PreviousFrom, truncate(request, 80), result.BackupDir)); err != nil {
		Fail(fmt.Sprintf("error registrando history: %s", err))
	}

	PrintSuccess(fmt.Sprintf("Ciclo reiniciado. Fase: %s → DISCOVERY", result.PreviousFrom))
	PrintInfo(fmt.Sprintf("Backup: %s", result.BackupDir))
	PrintInfo(fmt.Sprintf("Proyecto: %s", result.State.ProjectName))
	fmt.Println()
	printDiscoveryChatGuidance()
}
