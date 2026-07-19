package cmd

import (
	"fmt"
	"strings"

	"shipwright/pkg/harness"
)

func Start(args []string) {
	EnsureHarness()

	if len(args) == 0 {
		Fail("usage: shipwright start \"<request>\"")
	}

	request := strings.Join(args, " ")
	if strings.TrimSpace(request) == "" {
		Fail("la petición no puede estar vacía.")
	}

	state, err := harness.LoadState()
	if err != nil {
		Fail(err.Error())
	}

	if err := harness.StartRequest(state, request); err != nil {
		Fail(err.Error())
	}

	if err := state.Save(); err != nil {
		Fail(fmt.Sprintf("error guardando estado: %s", err))
	}

	nextAction := "Product Owner discovery round required. Open OpenCode, run /shipwright-active-agent, let product-owner ask discovery questions in chat, then generate product/context.md, product/assumptions.md, product/open-questions.md and product/scope.md."
	if err := harness.UpdateCurrent(state, nextAction); err != nil {
		Fail(fmt.Sprintf("error actualizando progress: %s", err))
	}

	if err := harness.AppendHistory("start", harness.StateDiscovery,
		fmt.Sprintf("Petición registrada: %s", truncate(request, 80))); err != nil {
		Fail(fmt.Sprintf("error registrando history: %s", err))
	}

	PrintSuccess("Petición registrada. Fase: INTAKE → DISCOVERY")
	PrintInfo(fmt.Sprintf("Proyecto: %s", state.ProjectName))
	fmt.Println()
	printDiscoveryChatGuidance()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
