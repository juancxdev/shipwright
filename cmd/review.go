package cmd

import (
	"fmt"

	"shipwright/pkg/harness"
)

func Review(args []string) {
	EnsureHarness()

	if len(args) == 0 {
		PrintUsage()
		Fail("usage: shipwright review <start|status>")
	}

	switch args[0] {
	case "start":
		reviewStart()
	case "status":
		reviewStatus()
	default:
		Fail(fmt.Sprintf("unknown review subcommand: %s\n\nValid: start | status", args[0]))
	}
}

func reviewStart() {
	state, err := harness.LoadState()
	if err != nil {
		Fail(err.Error())
	}

	generated, skipped, errors := harness.StartReviewArtifacts(state)

	fmt.Println("Shipwright — Review Start")
	fmt.Println("===========================")
	fmt.Println()

	if len(generated) > 0 {
		fmt.Println("Generated:")
		for _, path := range generated {
			fmt.Printf("  ✓ %s\n", path)
		}
	}

	if len(skipped) > 0 {
		fmt.Println()
		fmt.Println("Already exists:")
		for _, path := range skipped {
			fmt.Printf("  → %s\n", path)
		}
	}

	if len(errors) > 0 {
		fmt.Println()
		fmt.Println("Errors:")
		for _, e := range errors {
			fmt.Printf("  ✗ %s\n", e)
		}
		Exit(1)
		return
	}

	fmt.Println()
	PrintInfo("Fill these reports with real evidence. Placeholders block progress.")
	fmt.Println("Then run: shipwright review status")
}

func reviewStatus() {
	assessment := harness.AssessReviewReports()
	fmt.Print(harness.FormatReviewAssessment(assessment))
	if assessment.BlocksProgress() {
		Exit(1)
	}
}
