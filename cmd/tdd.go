package cmd

import (
	"fmt"
	"strings"

	"shipwright/pkg/harness"
)

func TDD(args []string) {
	EnsureHarness()
	if len(args) == 0 {
		printTDDUsage()
		return
	}

	switch args[0] {
	case "refresh":
		policy, err := harness.RefreshTDDPolicy()
		if err != nil {
			Fail(fmt.Sprintf("error refrescando TDD policy: %s", err))
		}
		PrintSuccess(fmt.Sprintf("TDD policy actualizada (%s)", policy.Mode))
	case "status":
		fmt.Print(harness.FormatTDDAssessment(harness.AssessTDDCompliance()))
	case "policy":
		policy, err := harness.LoadTDDPolicy()
		if err != nil {
			Fail(fmt.Sprintf("no se pudo leer TDD policy; ejecutá 'shipwright tdd refresh': %s", err))
		}
		fmt.Print(harness.RenderTDDPolicyMarkdown(policy))
	case "help", "-h", "--help":
		printTDDUsage()
	default:
		Fail(fmt.Sprintf("subcomando tdd desconocido: %s\n\n%s", args[0], strings.TrimSpace(tddUsageText())))
	}
}

func printTDDUsage() {
	fmt.Print(tddUsageText())
}

func tddUsageText() string {
	return `Shipwright — TDD Harness

Usage:
  shipwright tdd refresh   Regenerate .harness/tdd-policy.* from project calibration
  shipwright tdd status    Show strict/suggested/none mode and evidence status
  shipwright tdd policy    Print the current TDD policy markdown
`
}
