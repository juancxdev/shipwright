package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"shipwright/pkg/harness"
)

func Config(args []string) {
	EnsureHarness()

	if len(args) == 0 {
		PrintUsage()
		Fail("usage: shipwright config <show|init|env|validate>")
	}

	switch args[0] {
	case "show":
		configShow(args[1:])
	case "init":
		configInit(args[1:])
	case "env":
		configEnv(args[1:])
	case "validate":
		configValidate(args[1:])
	default:
		Fail(fmt.Sprintf("unknown config subcommand: %s\n\nValid: show | init | env | validate", args[0]))
	}
}

func configShow(args []string) {
	probe := harness.RealSystemProbe{}
	cfg, err := harness.LoadEffectivePortableConfig(probe)
	if err != nil {
		Fail(fmt.Sprintf("cannot load portable config: %s", err))
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		Fail(fmt.Sprintf("cannot render config: %s", err))
	}
	fmt.Println(string(data))
}

func configInit(args []string) {
	if _, err := os.Stat(harness.PortableConfigFile); err == nil {
		PrintInfo("Portable config already exists: " + harness.PortableConfigFile)
		return
	} else if !errors.Is(err, os.ErrNotExist) {
		Fail(fmt.Sprintf("cannot check portable config: %s", err))
	}

	cfg := harness.DefaultPortableConfig()
	if err := cfg.Save(); err != nil {
		Fail(fmt.Sprintf("cannot write portable config: %s", err))
	}
	PrintSuccess("Portable config created: " + harness.PortableConfigFile)
}

func configEnv(args []string) {
	fmt.Println("Shipwright — Portable Configuration Environment Overrides")
	fmt.Println("============================================================")
	fmt.Println()
	fmt.Println("General:")
	fmt.Println("  SHIPWRIGHT_HEALTH_TIMEOUT_MS   Health-check timeout in milliseconds")
	fmt.Println()
	fmt.Println("Engram:")
	fmt.Println("  ENGRAM_BINARY       Absolute path to the Engram binary")
	fmt.Println("  ENGRAM_HEALTH_URL   Health endpoint for future Engram checks")
	fmt.Println()
	fmt.Println("OpenPencil:")
	fmt.Println("  OPENPENCIL_APP_PATH      Absolute path to OpenPencil app/install dir")
	fmt.Println("  OPENPENCIL_MCP_SERVER    Absolute path to legacy/bundled OpenPencil MCP JS server")
	fmt.Println("  OPENPENCIL_MCP_COMMAND   OpenPencil MCP command, usually openpencil-mcp")
	fmt.Println("  OPENPENCIL_CANVAS_ACTIVE Temporary signal used by detection tests until real MCP handshake exists")
	fmt.Println()
	fmt.Println("OpenCode models:")
	fmt.Println("  SHIPWRIGHT_OPENCODE_DEFAULT_MODEL     Fallback model for generated OpenCode agents")
	fmt.Println("  SHIPWRIGHT_OPENCODE_REASONING_MODEL   Model for reasoning-heavy roles")
	fmt.Println("  SHIPWRIGHT_OPENCODE_FAST_MODEL        Model for lighter/documentary roles")
	fmt.Println("  SHIPWRIGHT_OPENCODE_AGENT_MODELS      Comma list: role=model,role=model")
}

func configValidate(args []string) {
	jsonOutput := false
	for _, arg := range args {
		switch arg {
		case "--json":
			jsonOutput = true
		case "-h", "--help":
			fmt.Println("usage: shipwright config validate [--json]")
			return
		default:
			Fail(fmt.Sprintf("unknown config validate option: %s\n\nValid: --json", arg))
		}
	}

	cfg, err := harness.LoadPortableConfigRaw()
	if err != nil {
		Fail(fmt.Sprintf("cannot load portable config: %s", err))
	}
	issues := harness.ValidatePortableConfig(cfg)
	errorsCount, warningsCount := harness.CountConfigIssueSeverities(issues)

	if jsonOutput {
		payload := struct {
			Valid    bool                            `json:"valid"`
			Errors   int                             `json:"errors"`
			Warnings int                             `json:"warnings"`
			Issues   []harness.ConfigValidationIssue `json:"issues"`
		}{
			Valid:    errorsCount == 0,
			Errors:   errorsCount,
			Warnings: warningsCount,
			Issues:   issues,
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			Fail(fmt.Sprintf("cannot render config validation: %s", err))
		}
		fmt.Println(string(data))
		if errorsCount > 0 {
			Exit(2)
		}
		return
	}

	fmt.Println("Shipwright — Config Validation")
	fmt.Println("=================================")
	fmt.Println()
	if len(issues) == 0 {
		PrintSuccess("Portable config is valid.")
		return
	}
	for _, issue := range issues {
		fmt.Printf("[%s] %s — %s\n", issue.Severity, issue.Path, issue.Message)
		if issue.Action != "" {
			fmt.Printf("     action: %s\n", issue.Action)
		}
	}
	fmt.Println()
	fmt.Printf("Summary: %d error(s), %d warning(s)\n", errorsCount, warningsCount)
	if errorsCount > 0 {
		Exit(2)
	}
}
