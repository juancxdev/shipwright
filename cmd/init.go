package cmd

import (
	"fmt"
	"os"
	"strings"

	"shipwright/pkg/harness"
)

func Init(args []string) {
	initOptions := parseInitOptions(args)
	executorName := initOptions.Executor
	if harnessInitialized() {
		Fail("el harness ya está inicializado en este directorio.")
	}

	cwd, err := os.Getwd()
	if err != nil {
		Fail(fmt.Sprintf("no se pudo obtener el directorio actual: %s", err))
	}

	projectName := filepathBase(cwd)

	profile, err := harness.CalibrateProject(projectName)
	if err != nil {
		Fail(fmt.Sprintf("error calibrando proyecto: %s", err))
	}

	fmt.Printf("Inicializando Shipwright en: %s\n", cwd)
	fmt.Printf("Proyecto: %s\n\n", projectName)

	if err := harness.CreateBaseStructure(); err != nil {
		Fail(fmt.Sprintf("error creando estructura: %s", err))
	}
	PrintSuccess("Estructura de carpetas creada")

	roles := harness.ListRoleNames()
	if err := harness.WriteRoles(); err != nil {
		Fail(fmt.Sprintf("error escribiendo roles: %s", err))
	}
	PrintSuccess(fmt.Sprintf("Roles de agentes definidos (%d)", len(roles)))

	portableConfig := harness.DefaultPortableConfig()
	if initOptions.OpenCodeModels.Used {
		harness.ApplyOpenCodeModelOverrides(portableConfig, initOptions.OpenCodeModels.Overrides)
	}
	if err := portableConfig.Save(); err != nil {
		Fail(fmt.Sprintf("error guardando config.json: %s", err))
	}
	PrintSuccess("Portable config creada (.harness/config.json)")
	if initOptions.OpenCodeModels.Used {
		PrintSuccess("OpenCode model config aplicada (.harness/config.json)")
	}

	integrations := harness.DefaultIntegrations()
	integrations.ApplyPortableConfig(portableConfig)
	if err := integrations.Save(); err != nil {
		Fail(fmt.Sprintf("error guardando integrations.json: %s", err))
	}
	PrintSuccess("Integrations state creado (.harness/integrations.json)")

	state := harness.NewState(projectName)
	if err := state.Save(); err != nil {
		Fail(fmt.Sprintf("error guardando state.json: %s", err))
	}
	PrintSuccess("State inicial creado (.harness/state.json)")

	if err := harness.InitProgress(); err != nil {
		Fail(fmt.Sprintf("error inicializando progress: %s", err))
	}
	PrintSuccess("Progress log inicializado (progress/current.md, progress/history.md)")

	if err := harness.SaveProjectProfile(profile); err != nil {
		Fail(fmt.Sprintf("error guardando project profile: %s", err))
	}
	PrintSuccess("Project calibration creada (.harness/project-profile.json, .harness/project-profile.md)")
	printProjectCalibrationSummary(profile)

	tddPolicy, err := harness.RefreshTDDPolicyFromProfile(profile)
	if err != nil {
		Fail(fmt.Sprintf("error guardando TDD policy: %s", err))
	}
	PrintSuccess(fmt.Sprintf("TDD policy creada (.harness/tdd-policy.json, mode=%s)", tddPolicy.Mode))

	if executorName != "" {
		result, err := harness.GenerateExecutor(executorName)
		if err != nil {
			Fail(fmt.Sprintf("error generando executor %s: %s", executorName, err))
		}
		PrintSuccess(fmt.Sprintf("Executor %s generado (%d creados, %d actualizados)", result.Name, len(result.FilesCreated), len(result.FilesUpdated)))
	}

	registry, err := harness.RefreshSkillRegistry()
	if err != nil {
		Fail(fmt.Sprintf("error refrescando skill registry: %s", err))
	}
	PrintSuccess(fmt.Sprintf("Skill registry creado (.harness/skill-registry.json, %d skills)", len(registry.Skills)))
	digests, err := harness.RefreshSkillDigestsFromRegistry(registry)
	if err != nil {
		Fail(fmt.Sprintf("error generando skill digests: %s", err))
	}
	PrintSuccess(fmt.Sprintf("Skill digests creados (.harness/skill-digests.json, %d agentes)", len(digests.Digests)))

	fmt.Println()
	fmt.Println("Estructura creada:")
	fmt.Println("  .harness/          — estado, project profile, TDD policy, skill registry, skill digests, agentes, approvals, integrations")
	fmt.Println("  product/           — discovery, contexto, alcance")
	fmt.Println("  project/           — planificación PMBOK-lite")
	fmt.Println("  design/            — UX/UI (OpenPencil-ready)")
	fmt.Println("  architecture/      — decisiones técnicas")
	fmt.Println("  contracts/         — OpenAPI, eventos")
	fmt.Println("  backlog/           — epics, stories, tasks")
	fmt.Println("  sdd/               — specs, design, tasks")
	fmt.Println("  knowledge/         — conocimiento reusable")
	fmt.Println("  progress/          — current.md, history.md")
	fmt.Println("  reports/           — QA, security, contract tests")
	if executorName != "" {
		fmt.Printf("  executor: %-9s — AI executor bootstrap generated\n", executorName)
	}
	fmt.Println()
	fmt.Println("Próximo paso:")
	fmt.Println("  shipwright start \"tu petición de producto\"")
	if executorName == harness.ExecutorOpenCode {
		fmt.Println("  o simplemente ejecutá: opencode")
	}
}

type initOptions struct {
	Executor       string
	OpenCodeModels openCodeModelFlagParseResult
}

const defaultInitExecutor = harness.ExecutorOpenCode

func parseInitOptions(args []string) initOptions {
	options := initOptions{
		Executor:       defaultInitExecutor,
		OpenCodeModels: parseOpenCodeModelFlags(args),
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if (arg == "--executor" || arg == "--ai") && i+1 < len(args) {
			options.Executor = args[i+1]
			i++
			continue
		}
		if len(arg) > len("--executor=") && arg[:len("--executor=")] == "--executor=" {
			options.Executor = arg[len("--executor="):]
		}
		if len(arg) > len("--ai=") && arg[:len("--ai=")] == "--ai=" {
			options.Executor = arg[len("--ai="):]
		}
	}
	return options
}

func filepathBase(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == os.PathSeparator {
			return path[i+1:]
		}
	}
	return path
}

func printProjectCalibrationSummary(profile *harness.ProjectProfile) {
	if profile == nil {
		return
	}
	fmt.Println()
	fmt.Println("Calibración del proyecto:")
	if len(profile.Languages) == 0 {
		fmt.Println("  stack:      no detectado (greenfield)")
	} else {
		fmt.Printf("  stack:      %s\n", strings.Join(profile.Languages, ", "))
	}
	if len(profile.Commands.Test) == 0 {
		fmt.Println("  tests:      no detectado")
	} else {
		fmt.Printf("  tests:      %s\n", profile.Commands.Test[0].Command)
	}
	fmt.Printf("  TDD mode:   %s\n", profile.TDD.RecommendedMode)
	if len(profile.Warnings) > 0 {
		fmt.Printf("  warnings:   %d\n", len(profile.Warnings))
	}
}
