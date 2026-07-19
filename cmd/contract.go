package cmd

import (
	"fmt"

	"shipwright/pkg/harness"
)

func Contract(args []string) {
	EnsureHarness()

	if len(args) == 0 {
		PrintUsage()
		Fail("usage: shipwright contract <validate|generate-tasks|check-mocks|check-compliance|show>")
	}

	subcommand := args[0]
	rest := args[1:]

	switch subcommand {
	case "validate":
		contractValidate(rest)
	case "generate-tasks":
		contractGenerateTasks(rest)
	case "check-mocks":
		contractCheckMocks(rest)
	case "check-compliance":
		contractCheckCompliance(rest)
	case "show":
		contractShow(rest)
	default:
		Fail(fmt.Sprintf("unknown contract subcommand: %s\n\nValid: validate | generate-tasks | check-mocks | check-compliance | show", subcommand))
	}
}

func contractValidate(args []string) {
	if !harness.ContractExists() {
		Fail("contracts/openapi.yaml does not exist. Run: shipwright generate contracts/openapi.yaml")
	}

	result := harness.ValidateContract(harness.ContractFile)

	fmt.Println("Shipwright — Contract Validation")
	fmt.Println("====================================")
	fmt.Println()

	if len(result.Errors) > 0 {
		fmt.Println("Errors:")
		for _, e := range result.Errors {
			fmt.Printf("  ✗ %s\n", e)
		}
		fmt.Println()
	}

	if len(result.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, w := range result.Warnings {
			fmt.Printf("  ⚠ %s\n", w)
		}
		fmt.Println()
	}

	fmt.Printf("Endpoints: %d\n", result.Endpoints)
	fmt.Printf("Schemas:   %d\n", result.Schemas)

	if result.IsValid {
		fmt.Println()
		PrintSuccess("Contract is valid.")
	} else {
		fmt.Println()
		Fail("Contract validation FAILED.")
	}
}

func contractGenerateTasks(args []string) {
	if !harness.ContractExists() {
		Fail("contracts/openapi.yaml does not exist. Generate it first or run: shipwright scaffold")
	}

	parseResult := harness.ParseContract(harness.ContractFile)
	if !parseResult.IsValid {
		Fail("Contract is invalid. Run: shipwright contract validate")
	}

	spec := parseResult.Spec

	state, err := harness.LoadState()
	if err != nil {
		Fail(err.Error())
	}

	feTasks, beTasks := harness.GenerateContractTasks(spec, state.ProjectName)

	fePath := "backlog/frontend-tasks.md"
	bePath := "backlog/backend-tasks.md"

	if harness.ArtifactExists(fePath) {
		fmt.Printf("  → %s (already exists, overwriting)\n", fePath)
	}
	if err := harness.WriteFile(fePath, feTasks); err != nil {
		Fail(fmt.Sprintf("cannot write %s: %s", fePath, err))
	}
	PrintSuccess(fmt.Sprintf("Generated: %s (%d endpoints)", fePath, spec.EndpointCount))

	if harness.ArtifactExists(bePath) {
		fmt.Printf("  → %s (already exists, overwriting)\n", bePath)
	}
	if err := harness.WriteFile(bePath, beTasks); err != nil {
		Fail(fmt.Sprintf("cannot write %s: %s", bePath, err))
	}
	PrintSuccess(fmt.Sprintf("Generated: %s (%d endpoints)", bePath, spec.EndpointCount))

	fmt.Println()
	fmt.Println("Contract-first rules enforced in generated tasks:")
	fmt.Println("  Frontend: mock mode mandatory, HTTP adapter separate, no mock deletion")
	fmt.Println("  Backend:  API must match contract, errors consistent, no silent contract changes")
}

func contractCheckMocks(args []string) {
	if !harness.ContractExists() {
		Fail("contracts/openapi.yaml does not exist.")
	}

	parseResult := harness.ParseContract(harness.ContractFile)
	if parseResult.Spec == nil {
		Fail("cannot parse contract.")
	}

	if !harness.ArtifactExists("progress/frontend.md") {
		Fail("progress/frontend.md does not exist. Run: shipwright scaffold or shipwright agents run frontend-engineer")
	}

	result := harness.CheckMockCompliance(parseResult.Spec)
	fmt.Print(harness.FormatMockCompliance(result))

	if len(result.Issues) > 0 {
		Exit(1)
	}
}

func contractCheckCompliance(args []string) {
	if !harness.ContractExists() {
		Fail("contracts/openapi.yaml does not exist.")
	}

	parseResult := harness.ParseContract(harness.ContractFile)
	if parseResult.Spec == nil {
		Fail("cannot parse contract.")
	}

	if !harness.ArtifactExists("progress/backend.md") {
		Fail("progress/backend.md does not exist. Run: shipwright scaffold or shipwright agents run backend-engineer")
	}

	result := harness.CheckBackendCompliance(parseResult.Spec)
	fmt.Print(harness.FormatBackendCompliance(result))

	if len(result.Issues) > 0 {
		Exit(1)
	}
}

func contractShow(args []string) {
	if !harness.ContractExists() {
		Fail("contracts/openapi.yaml does not exist.")
	}

	parseResult := harness.ParseContract(harness.ContractFile)
	if parseResult.Spec == nil {
		Fail("cannot parse contract.")
	}

	fmt.Println("Shipwright — Contract Summary")
	fmt.Println("================================")
	fmt.Println()

	fmt.Print(harness.FormatContractSpec(parseResult.Spec))

	if len(parseResult.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, w := range parseResult.Warnings {
			fmt.Printf("  ⚠ %s\n", w)
		}
	}

	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  shipwright contract validate        — validate contract structure")
	fmt.Println("  shipwright contract generate-tasks  — generate FE/BE tasks from contract")
	fmt.Println("  shipwright contract check-mocks     — verify frontend mock compliance")
	fmt.Println("  shipwright contract check-compliance— verify backend contract compliance")
}
