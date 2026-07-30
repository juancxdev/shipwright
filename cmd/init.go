package cmd

import (
	"fmt"
	"os"
	"strings"

	"shipwright/internal/app/harness"
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

	if err := harness.EnsureCommunicationPolicy(); err != nil {
		Fail(fmt.Sprintf("error creando communication policy: %s", err))
	}
	PrintSuccess("Communication policy creada (.harness/communication-policy.md)")

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

	if initOptions.IntegrationWizard && shouldRunInitIntegrationWizard() {
		if err := runInitIntegrationWizard(portableConfig, integrations); err != nil {
			Fail(fmt.Sprintf("error configurando integraciones: %s", err))
		}
		PrintSuccess("Integraciones calibradas (.harness/integrations.json)")
	} else if initOptions.IntegrationWizard {
		PrintInfo("Integration wizard omitido (entorno no interactivo o CI).")
	} else {
		PrintInfo("Integration wizard omitido por flag.")
	}

	state := harness.NewState(projectName)
	if err := state.Save(); err != nil {
		Fail(fmt.Sprintf("error guardando state.json: %s", err))
	}
	PrintSuccess("State inicial creado (.harness/state.json)")

	if err := harness.InitProgress(); err != nil {
		Fail(fmt.Sprintf("error inicializando progress: %s", err))
	}
	PrintSuccess("Progress log inicializado (.harness/artifacts/progress/current.md, .harness/artifacts/progress/history.md)")

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

	modelPolicy := harness.DefaultModelPolicy(portableConfig.Executors.OpenCode)
	if err := harness.SaveModelPolicy(modelPolicy); err != nil {
		Fail(fmt.Sprintf("error guardando model policy: %s", err))
	}
	PrintSuccess("Model policy creada (.harness/model-policy.json)")

	if executorName != "" {
		result, err := harness.GenerateExecutor(executorName)
		if err != nil {
			Fail(fmt.Sprintf("error generando executor %s: %s", executorName, err))
		}
		PrintSuccess(fmt.Sprintf("Executor %s generado (%d creados, %d actualizados)", result.Name, len(result.FilesCreated), len(result.FilesUpdated)))
	}

	manifest, err := harness.SaveRecommendedSkillPackManifest(profile)
	if err != nil {
		Fail(fmt.Sprintf("error guardando skill pack manifest: %s", err))
	}
	PrintSuccess(fmt.Sprintf("Skill pack manifest creado (.harness/skill-packs.json, %d recomendados)", len(manifest.Recommended)))

	registry, err := harness.RefreshSkillRegistry()
	if err != nil {
		Fail(fmt.Sprintf("error refrescando skill registry: %s", err))
	}
	PrintSuccess(fmt.Sprintf("Skill registry creado (.harness/skill-registry.json, %d skills)", len(registry.Skills)))
	assignments, err := harness.RefreshSkillAssignmentsFromRegistry(registry)
	if err != nil {
		Fail(fmt.Sprintf("error generando skill assignments: %s", err))
	}
	PrintSuccess(fmt.Sprintf("Skill assignments creados (.harness/skill-assignments.json, %d tecnologías)", len(assignments.Technologies)))
	digests, err := harness.RefreshSkillDigestsFromRegistry(registry)
	if err != nil {
		Fail(fmt.Sprintf("error generando skill digests: %s", err))
	}
	PrintSuccess(fmt.Sprintf("Skill digests creados (.harness/skill-digests.json, %d agentes)", len(digests.Digests)))

	fmt.Println()
	fmt.Println("Estructura creada:")
	fmt.Println("  .harness/          — estado, communication policy, project profile, TDD policy, model policy, skill registry, skill packs, skill lock, agentes, approvals, integrations")
	fmt.Println("  .harness/artifacts/product/           — discovery, contexto, alcance")
	fmt.Println("  .harness/artifacts/project/           — planificación PMBOK-lite")
	fmt.Println("  .harness/artifacts/design/            — UX/UI (Stitch-first, OpenPencil optional)")
	fmt.Println("  .harness/artifacts/architecture/      — decisiones técnicas")
	fmt.Println("  .harness/artifacts/contracts/         — OpenAPI, eventos")
	fmt.Println("  .harness/artifacts/backlog/           — epics, stories, tasks")
	fmt.Println("  .harness/artifacts/sdd/               — specs, design, tasks")
	fmt.Println("  .harness/artifacts/knowledge/         — conocimiento reusable")
	fmt.Println("  .harness/artifacts/progress/          — current.md, history.md")
	fmt.Println("  .harness/artifacts/reports/           — QA, security, contract tests")
	if executorName != "" {
		fmt.Printf("  executor: %-9s — AI executor bootstrap generated\n", executorName)
	}
	fmt.Println()
	fmt.Println("Próximo paso:")
	fmt.Println("  shipwright start \"tu petición de producto\"")
	if executorName == harness.ExecutorOpenCode {
		fmt.Println("  o simplemente ejecuta: opencode")
	}
	printOptionalSkillImportHint(profile, assignments)
}

type initOptions struct {
	Executor          string
	OpenCodeModels    openCodeModelFlagParseResult
	IntegrationWizard bool
}

const defaultInitExecutor = harness.ExecutorOpenCode

func parseInitOptions(args []string) initOptions {
	options := initOptions{
		Executor:          defaultInitExecutor,
		OpenCodeModels:    parseOpenCodeModelFlags(args),
		IntegrationWizard: true,
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
		if arg == "--no-interactive" || arg == "--no-integrations-wizard" {
			options.IntegrationWizard = false
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

func printOptionalSkillImportHint(profile *harness.ProjectProfile, assignments *harness.SkillAssignmentSet) {
	fmt.Println()
	fmt.Println("Skills opcionales:")
	if profile != nil && (len(profile.Languages) > 0 || len(profile.Stacks) > 0) {
		var detected []string
		detected = append(detected, profile.Languages...)
		for _, stack := range profile.Stacks {
			detected = append(detected, stack.Name)
		}
		fmt.Printf("  tecnologías detectadas: %s\n", strings.Join(harness.SortedUniqueForDisplay(detected), ", "))
	} else {
		fmt.Println("  tecnologías detectadas: greenfield/no detectado todavía")
	}
	if assignments != nil && len(assignments.Skills) > 0 {
		fmt.Printf("  recomendaciones actuales: %d skills\n", len(assignments.Skills))
	}
	fmt.Println("  Shipwright no ejecuta instaladores externos durante init.")
	if harness.AutoSkillsAvailable() {
		fmt.Println("  Se detectó .agents/skills. Puedes importarlas con:")
		fmt.Println("    shipwright skills import autoskills")
	} else {
		fmt.Println("  Si usas autoskills después, importa sus skills con:")
		fmt.Println("    shipwright skills import autoskills")
	}
}
