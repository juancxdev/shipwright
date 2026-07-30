package cmd

import (
	"flag"
	"fmt"
	"strings"

	"shipwright/internal/app/harness"
)

func Integrations(args []string) {
	EnsureHarness()

	if len(args) == 0 {
		PrintUsage()
		Fail("usage: shipwright integrations <status|enable|disable|detect|configure>")
	}

	subcommand := args[0]
	rest := args[1:]

	switch subcommand {
	case "status":
		integrationsStatus(rest)
	case "enable":
		integrationsEnable(rest)
	case "disable":
		integrationsDisable(rest)
	case "detect":
		integrationsDetect(rest)
	case "configure":
		integrationsConfigure(rest)
	default:
		Fail(fmt.Sprintf("unknown integrations subcommand: %s\n\nValid: status | enable | disable | detect | configure", subcommand))
	}
}

func integrationsStatus(args []string) {
	integrations, err := harness.LoadIntegrations()
	if err != nil {
		Fail(fmt.Sprintf("cannot load integrations: %s", err))
	}

	fmt.Println("Shipwright — Integrations Status")
	fmt.Println("=====================================")
	fmt.Println()

	if integrations.Platform.OS != "" {
		fmt.Println("Platform:")
		fmt.Printf("  os:       %s\n", integrations.Platform.OS)
		fmt.Printf("  arch:     %s\n", integrations.Platform.Arch)
		fmt.Printf("  ci:       %s\n", boolYesNo(integrations.Platform.IsCI))
		fmt.Println()
	}

	fmt.Println("Engram (Memory):")
	fmt.Printf("  enabled:  %s\n", boolEnabled(integrations.Engram.Enabled))
	fmt.Printf("  mode:     %s\n", integrations.Engram.Mode)
	fmt.Printf("  status:   %s\n", integrations.Engram.Status)
	fmt.Printf("  fallback: %s\n", integrations.Engram.Fallback)
	if integrations.Engram.BinaryPath != "" {
		fmt.Printf("  binary:   %s\n", integrations.Engram.BinaryPath)
	}
	if integrations.Engram.Reason != "" {
		fmt.Printf("  reason:   %s\n", integrations.Engram.Reason)
	}
	fmt.Println()

	fmt.Println("Stitch (Design primary):")
	fmt.Printf("  enabled:  %s\n", boolEnabled(integrations.Stitch.Enabled))
	fmt.Printf("  mode:     %s\n", integrations.Stitch.Mode)
	fmt.Printf("  status:   %s\n", integrations.Stitch.Status)
	fmt.Printf("  fallback: %s\n", integrations.Stitch.Fallback)
	if integrations.Stitch.Reason != "" {
		fmt.Printf("  reason:   %s\n", integrations.Stitch.Reason)
	}
	fmt.Println()

	fmt.Println("OpenDesign (Design MCP optional):")
	fmt.Printf("  enabled:  %s\n", boolEnabled(integrations.OpenDesign.Enabled))
	fmt.Printf("  mode:     %s\n", integrations.OpenDesign.Mode)
	fmt.Printf("  status:   %s\n", integrations.OpenDesign.Status)
	fmt.Printf("  fallback: %s\n", integrations.OpenDesign.Fallback)
	if integrations.OpenDesign.MCPCommand != "" {
		fmt.Printf("  command:  %s\n", integrations.OpenDesign.MCPCommand)
	}
	if len(integrations.OpenDesign.MCPArgs) > 0 {
		fmt.Printf("  args:     %s\n", strings.Join(integrations.OpenDesign.MCPArgs, " "))
	}
	if integrations.OpenDesign.DaemonURL != "" {
		fmt.Printf("  daemon:   %s\n", integrations.OpenDesign.DaemonURL)
	}
	if integrations.OpenDesign.DataDir != "" {
		fmt.Printf("  data dir: %s\n", integrations.OpenDesign.DataDir)
	}
	if integrations.OpenDesign.IPCPath != "" {
		fmt.Printf("  ipc:      %s\n", integrations.OpenDesign.IPCPath)
	}
	if integrations.OpenDesign.Reason != "" {
		fmt.Printf("  reason:   %s\n", integrations.OpenDesign.Reason)
	}
	fmt.Println()

	fmt.Println("OpenPencil (Design optional):")
	fmt.Printf("  enabled:  %s\n", boolEnabled(integrations.OpenPencil.Enabled))
	fmt.Printf("  mode:     %s\n", integrations.OpenPencil.Mode)
	fmt.Printf("  status:   %s\n", integrations.OpenPencil.Status)
	fmt.Printf("  fallback: %s\n", integrations.OpenPencil.Fallback)
	if integrations.OpenPencil.AppPath != "" {
		fmt.Printf("  app:      %s\n", integrations.OpenPencil.AppPath)
	}
	if integrations.OpenPencil.MCPServerPath != "" {
		fmt.Printf("  mcp:      %s\n", integrations.OpenPencil.MCPServerPath)
	}
	if integrations.OpenPencil.MCPCommand != "" {
		fmt.Printf("  command:  %s\n", integrations.OpenPencil.MCPCommand)
	}
	if integrations.OpenPencil.Reason != "" {
		fmt.Printf("  reason:   %s\n", integrations.OpenPencil.Reason)
	}
	fmt.Println()

	if integrations.IsEngramEnabled() {
		adapter := harness.NewEngramMemoryAdapter()
		total, pending, synced := adapter.Stats()
		fmt.Printf("Engram queue: %d total, %d pending, %d synced\n", total, pending, synced)
	} else {
		localCount := harness.CountLocalEntries()
		fmt.Printf("Engram local entries: %d in %s\n", localCount, harness.DecisionsFile)
	}
	fmt.Println()

	mode, fallbackUsed, _ := harness.LoadDesignState()
	if mode != "" {
		fmt.Printf("Design mode: %s", mode)
		if fallbackUsed {
			fmt.Print(" (fallback was used)")
		}
		fmt.Println()
	} else {
		fmt.Println("Design mode: (not started)")
	}
}

func integrationsEnable(args []string) {
	if len(args) == 0 {
		Fail("usage: shipwright integrations enable <engram|stitch|opendesign|openpencil>")
	}

	target := args[0]
	integrations, err := harness.LoadIntegrations()
	if err != nil {
		Fail(fmt.Sprintf("cannot load integrations: %s", err))
	}

	switch target {
	case "engram":
		if integrations.IsEngramEnabled() {
			PrintInfo("Engram already enabled.")
			return
		}
		integrations.EnableEngram()
		if err := integrations.Save(); err != nil {
			Fail(fmt.Sprintf("cannot save: %s", err))
		}
		PrintSuccess("Engram enabled. Events will queue in .harness/memory-queue.json")

	case "stitch":
		if integrations.IsStitchEnabled() {
			PrintInfo("Stitch already enabled.")
			return
		}
		integrations.EnableStitch()
		if err := integrations.Save(); err != nil {
			Fail(fmt.Sprintf("cannot save: %s", err))
		}
		PrintSuccess("Stitch enabled. Design will use Google Stitch as primary provider.")
		PrintInfo("Set STITCH_API_KEY, or STITCH_ACCESS_TOKEN + GOOGLE_CLOUD_PROJECT, then run 'shipwright integrations detect'.")

	case "opendesign", "open-design":
		if integrations.IsOpenDesignEnabled() {
			PrintInfo("OpenDesign already enabled.")
			return
		}
		integrations.EnableOpenDesign()
		if err := integrations.Save(); err != nil {
			Fail(fmt.Sprintf("cannot save: %s", err))
		}
		PrintSuccess("OpenDesign enabled as optional MCP design provider.")
		PrintInfo("Run 'shipwright integrations configure opendesign ...' if command/data/ipc paths are not configured yet.")

	case "openpencil":
		if integrations.IsOpenPencilEnabled() {
			PrintInfo("OpenPencil already enabled.")
			return
		}
		integrations.EnableOpenPencil()
		if err := integrations.Save(); err != nil {
			Fail(fmt.Sprintf("cannot save: %s", err))
		}
		PrintSuccess("OpenPencil enabled as optional design provider. Stitch remains primary unless disabled or user explicitly requests OpenPencil.")
		PrintInfo("Run 'shipwright integrations detect' to check if OpenPencil is installed.")

	default:
		Fail(fmt.Sprintf("unknown integration: %s\n\nValid: engram | stitch | opendesign | openpencil", target))
	}
}

func integrationsDisable(args []string) {
	if len(args) == 0 {
		Fail("usage: shipwright integrations disable <engram|stitch|opendesign|openpencil>")
	}

	target := args[0]
	integrations, err := harness.LoadIntegrations()
	if err != nil {
		Fail(fmt.Sprintf("cannot load integrations: %s", err))
	}

	switch target {
	case "engram":
		if !integrations.IsEngramEnabled() {
			PrintInfo("Engram already disabled.")
			return
		}
		integrations.DisableEngram()
		if err := integrations.Save(); err != nil {
			Fail(fmt.Sprintf("cannot save: %s", err))
		}
		PrintSuccess("Engram disabled. Events will write to .harness/artifacts/progress/decisions.md")

	case "stitch":
		if !integrations.IsStitchEnabled() {
			PrintInfo("Stitch already disabled.")
			return
		}
		integrations.DisableStitch()
		if err := integrations.Save(); err != nil {
			Fail(fmt.Sprintf("cannot save: %s", err))
		}
		PrintSuccess("Stitch disabled. Design will use OpenPencil if enabled, otherwise doc-only mode.")

	case "opendesign", "open-design":
		if !integrations.IsOpenDesignEnabled() {
			PrintInfo("OpenDesign already disabled.")
			return
		}
		integrations.DisableOpenDesign()
		if err := integrations.Save(); err != nil {
			Fail(fmt.Sprintf("cannot save: %s", err))
		}
		PrintSuccess("OpenDesign disabled. Design will use Stitch/OpenPencil if enabled, otherwise doc-only mode.")

	case "openpencil":
		if !integrations.IsOpenPencilEnabled() {
			PrintInfo("OpenPencil already disabled.")
			return
		}
		integrations.DisableOpenPencil()
		if err := integrations.Save(); err != nil {
			Fail(fmt.Sprintf("cannot save: %s", err))
		}
		PrintSuccess("OpenPencil disabled. Design will use Stitch if enabled, otherwise doc-only mode.")

	default:
		Fail(fmt.Sprintf("unknown integration: %s\n\nValid: engram | stitch | opendesign | openpencil", target))
	}
}

func integrationsDetect(args []string) {
	integrations, err := harness.LoadIntegrations()
	if err != nil {
		Fail(fmt.Sprintf("cannot load integrations: %s", err))
	}

	fmt.Println("Shipwright — Integration Detection")
	fmt.Println("======================================")
	fmt.Println()

	probe := harness.RealSystemProbe{}
	config, err := harness.LoadEffectivePortableConfig(probe)
	if err != nil {
		Fail(fmt.Sprintf("cannot load portable config: %s", err))
	}
	integrations.ApplyPortableConfig(config)
	engram := harness.DetectEngramWithConfig(probe, config)
	stitch := harness.DetectStitchWithConfig(probe, config)
	opendesign := harness.DetectOpenDesignWithConfig(probe, config)
	openpencil := harness.DetectOpenPencilWithConfig(probe, config)
	integrations.ApplyDetection(engram, openpencil)
	integrations.ApplyStitchDetection(stitch)
	integrations.ApplyOpenDesignDetection(opendesign)

	engramInstalled := engram.Installed
	fmt.Println("Engram:")
	fmt.Printf("  MCP available: %s\n", boolYesNo(engramInstalled))
	fmt.Printf("  status:        %s\n", engram.Status)
	if engram.Path != "" {
		fmt.Printf("  path:          %s\n", engram.Path)
	}
	if engram.Reason != "" {
		fmt.Printf("  reason:        %s\n", engram.Reason)
	}
	if engramInstalled && !integrations.IsEngramEnabled() {
		fmt.Printf("  → recommendation: enable with 'shipwright integrations enable engram'\n")
	} else if !engramInstalled {
		fmt.Printf("  → fallback: %s\n", engram.Fallback)
	}
	fmt.Println()

	fmt.Println("Stitch:")
	fmt.Printf("  credentials:   %s\n", boolYesNo(stitch.Available))
	fmt.Printf("  status:        %s\n", stitch.Status)
	if stitch.Reason != "" {
		fmt.Printf("  reason:        %s\n", stitch.Reason)
	}
	if stitch.Available {
		fmt.Printf("  → design provider: stitch\n")
	} else {
		fmt.Printf("  → set STITCH_API_KEY or STITCH_ACCESS_TOKEN + GOOGLE_CLOUD_PROJECT\n")
	}
	fmt.Println()

	opInstalled := openpencil.Installed
	fmt.Println("OpenDesign (optional MCP):")
	fmt.Printf("  configured:   %s\n", boolYesNo(opendesign.Configured))
	fmt.Printf("  available:    %s\n", boolYesNo(opendesign.Available))
	fmt.Printf("  status:       %s\n", opendesign.Status)
	if opendesign.Path != "" {
		fmt.Printf("  command:      %s\n", opendesign.Path)
	}
	if opendesign.DaemonURL != "" {
		fmt.Printf("  daemon:       %s\n", opendesign.DaemonURL)
	}
	if opendesign.Reason != "" {
		fmt.Printf("  reason:       %s\n", opendesign.Reason)
	}
	if opendesign.Available {
		fmt.Printf("  → MCP provider: open-design\n")
	} else {
		fmt.Printf("  → configure with: shipwright integrations configure opendesign --command <node> --arg <cli.js> --arg mcp --data-dir <dir> --ipc-path <socket>\n")
	}
	fmt.Println()

	fmt.Println("OpenPencil (optional):")
	fmt.Printf("  MCP available:  %s\n", boolYesNo(opInstalled))
	fmt.Printf("  status:         %s\n", openpencil.Status)
	if openpencil.Path != "" {
		fmt.Printf("  mcp path:       %s\n", openpencil.Path)
	}
	if openpencil.Reason != "" {
		fmt.Printf("  reason:         %s\n", openpencil.Reason)
	}
	if opInstalled && openpencil.Active {
		fmt.Printf("  Canvas verified: yes\n")
		if !integrations.IsOpenPencilEnabled() {
			fmt.Printf("  → recommendation: enable with 'shipwright integrations enable openpencil'\n")
		}
	} else if opInstalled {
		fmt.Printf("  Canvas verified: no (CLI cannot verify live OpenPencil canvas)\n")
		if !integrations.IsOpenPencilEnabled() {
			fmt.Printf("  → recommendation: enable with 'shipwright integrations enable openpencil'\n")
		}
		fmt.Printf("  → next check: open OpenCode and run/ask: opencode mcp list; then use open-pencil_get_editor_state\n")
		fmt.Printf("  → fallback only if OpenCode cannot see open-pencil_* MCP tools: %s\n", openpencil.Fallback)
	} else {
		fmt.Printf("  → fallback: %s\n", openpencil.Fallback)
	}

	if err := integrations.Save(); err != nil {
		PrintInfo(fmt.Sprintf("warning: could not save detection results: %s", err))
	}
}

func boolYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

type repeatFlag []string

func (r *repeatFlag) String() string { return strings.Join(*r, ",") }
func (r *repeatFlag) Set(value string) error {
	*r = append(*r, value)
	return nil
}

func integrationsConfigure(args []string) {
	if len(args) == 0 {
		Fail("usage: shipwright integrations configure <opendesign> [flags]")
	}
	target := strings.ToLower(strings.TrimSpace(args[0]))
	switch target {
	case "opendesign", "open-design":
		configureOpenDesign(args[1:])
	default:
		Fail(fmt.Sprintf("unknown configurable integration: %s\n\nValid: opendesign", target))
	}
}

func configureOpenDesign(args []string) {
	flags := flag.NewFlagSet("opendesign", flag.ContinueOnError)
	flags.SetOutput(ioDiscard{})
	var mcpArgs repeatFlag
	command := flags.String("command", "", "OpenDesign MCP command, usually od or an absolute node path")
	flags.Var(&mcpArgs, "arg", "OpenDesign MCP argument; repeat for each argument")
	daemonURL := flags.String("daemon-url", "", "OpenDesign daemon HTTP URL, for example http://127.0.0.1:7377")
	dataDir := flags.String("data-dir", "", "OpenDesign OD_DATA_DIR")
	ipcPath := flags.String("ipc-path", "", "OpenDesign OD_SIDECAR_IPC_PATH")
	fromEnv := flags.Bool("from-env", false, "read OPENDESIGN_* / OD_* env vars")
	auto := flags.Bool("auto", false, "autodetect OpenDesign config; default when no manual flags are provided")
	help := flags.Bool("help", false, "show help")
	if err := flags.Parse(args); err != nil {
		Fail("usage: shipwright integrations configure opendesign [--auto|--from-env|--command <od-or-node> --arg <cli.js> --arg mcp]")
	}
	if *help {
		fmt.Println("usage: shipwright integrations configure opendesign [--auto|--from-env|--command <od-or-node> --arg <cli.js> --arg mcp]")
		fmt.Println()
		fmt.Println("Default:")
		fmt.Println("  shipwright integrations configure opendesign")
		fmt.Println("    Attempts autodetection from `od`, Node, OPENDESIGN_ROOT, OPENDESIGN_* and OD_* env vars.")
		fmt.Println()
		fmt.Println("Manual fallback:")
		fmt.Println("  shipwright integrations configure opendesign --command /opt/homebrew/bin/node --arg /path/to/open-design/apps/daemon/dist/cli.js --arg mcp --daemon-url http://127.0.0.1:7377 --data-dir /path/to/open-design/.od --ipc-path /tmp/open-design/ipc/default/daemon.sock")
		return
	}

	probe := harness.RealSystemProbe{}
	cfg, err := harness.LoadPortableConfig()
	if err != nil {
		Fail(fmt.Sprintf("cannot load portable config: %s", err))
	}
	if *fromEnv {
		cfg.ApplyEnv(probe)
	}
	if strings.TrimSpace(*command) != "" {
		cfg.Integrations.OpenDesign.MCPCommand = strings.TrimSpace(*command)
	}
	if len(mcpArgs) > 0 {
		cfg.Integrations.OpenDesign.MCPArgs = append([]string{}, mcpArgs...)
	}
	if strings.TrimSpace(*daemonURL) != "" {
		cfg.Integrations.OpenDesign.DaemonURL = strings.TrimRight(strings.TrimSpace(*daemonURL), "/")
	}
	if strings.TrimSpace(*dataDir) != "" {
		cfg.Integrations.OpenDesign.DataDir = strings.TrimSpace(*dataDir)
	}
	if strings.TrimSpace(*ipcPath) != "" {
		cfg.Integrations.OpenDesign.IPCPath = strings.TrimSpace(*ipcPath)
	}
	manualInput := strings.TrimSpace(*command) != "" || len(mcpArgs) > 0 || strings.TrimSpace(*daemonURL) != "" || strings.TrimSpace(*dataDir) != "" || strings.TrimSpace(*ipcPath) != ""
	if *auto || !manualInput {
		detected := harness.AutoConfigureOpenDesign(probe, cfg)
		if detected.Configured && strings.TrimSpace(cfg.Integrations.OpenDesign.MCPCommand) != "" {
			PrintSuccess("OpenDesign autodetectado")
		} else if !manualInput {
			Fail("No pude autodetectar OpenDesign. Instala el CLI `od`, configura OPENDESIGN_ROOT, o usa --command/--arg manual.")
		}
	}
	cfg.Integrations.OpenDesign.Mode = harness.ConfigModeMCP
	cfg.Integrations.OpenDesign.Fallback = "design-doc-only"
	if strings.TrimSpace(cfg.Integrations.OpenDesign.MCPCommand) == "" {
		Fail("OpenDesign necesita --command, `od` en PATH, OPENDESIGN_ROOT, o OPENDESIGN_MCP_COMMAND")
	}
	if len(cfg.Integrations.OpenDesign.MCPArgs) == 0 {
		Fail("OpenDesign necesita argumentos MCP; con `od` es: mcp; con Node es: <cli.js> mcp")
	}
	if err := cfg.Save(); err != nil {
		Fail(fmt.Sprintf("cannot save portable config: %s", err))
	}

	integrations, err := harness.LoadIntegrations()
	if err != nil {
		Fail(fmt.Sprintf("cannot load integrations: %s", err))
	}
	integrations.ApplyPortableConfig(cfg)
	integrations.EnableOpenDesign()
	detected := harness.DetectOpenDesignWithConfig(probe, cfg)
	integrations.ApplyOpenDesignDetection(detected)
	if err := integrations.Save(); err != nil {
		Fail(fmt.Sprintf("cannot save integrations: %s", err))
	}

	PrintSuccess("OpenDesign configurado y activado para este proyecto")
	fmt.Printf("  status:  %s\n", detected.Status)
	if detected.Reason != "" {
		fmt.Printf("  detail:  %s\n", detected.Reason)
	}
	PrintInfo("Regenerá OpenCode con: shipwright executor generate opencode")
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
