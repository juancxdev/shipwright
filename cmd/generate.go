package cmd

import (
	"fmt"

	"shipwright/pkg/harness"
)

func Generate(args []string) {
	EnsureHarness()

	if len(args) == 0 {
		PrintUsage()
		Fail("usage: shipwright generate <artifact>\n\nRun 'shipwright generate --list' to see available artifacts.")
	}

	if args[0] == "--list" || args[0] == "-l" {
		listArtifacts()
		return
	}

	artifactPath := args[0]

	state, err := harness.LoadState()
	if err != nil {
		Fail(err.Error())
	}

	if err := harness.ScaffoldArtifact(state, artifactPath); err != nil {
		Fail(err.Error())
	}

	PrintSuccess(fmt.Sprintf("Generated: %s", artifactPath))
	PrintInfo("This is a PLACEHOLDER. Fill in with real content before approval.")
}

func Scaffold(args []string) {
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

	result := harness.ScaffoldPhase(state)

	fmt.Print(harness.FormatScaffoldResult(result))

	if len(result.Errors) > 0 {
		Exit(1)
		return
	}

	if len(result.Generated) > 0 {
		fmt.Println()
		PrintInfo("Placeholders generated. Fill in with real content before approval.")
		fmt.Println("Then run: shipwright next")
	} else if len(result.Skipped) > 0 && len(result.Generated) == 0 {
		fmt.Println()
		fmt.Println("All artifacts for this phase already exist.")
		fmt.Println("Run: shipwright next")
	}
}

func listArtifacts() {
	fmt.Println("Available artifacts for generation:")
	fmt.Println()

	artifacts := harness.ListScaffoldableArtifacts()
	for _, a := range artifacts {
		fmt.Printf("  %s\n", a)
	}

	fmt.Println()
	fmt.Println("Usage: shipwright generate <artifact>")
	fmt.Println("       shipwright scaffold  (generates all artifacts for current phase)")
}
