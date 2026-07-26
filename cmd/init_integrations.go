package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"

	"shipwright/internal/app/harness"
)

type initIntegrationOption struct {
	Number      int
	Key         string
	Name        string
	Description string
	Recommended bool
	Detection   harness.DetectionResult
}

type initIntegrationSelectorState struct {
	Options  []initIntegrationOption
	Cursor   int
	Selected map[string]bool
}

type initSelectorKey int

const (
	initSelectorKeyUnknown initSelectorKey = iota
	initSelectorKeyUp
	initSelectorKeyDown
	initSelectorKeyToggle
	initSelectorKeySubmit
	initSelectorKeyCancel
)

func shouldRunInitIntegrationWizard() bool {
	probe := harness.RealSystemProbe{}
	if harness.DetectPlatform(probe).IsCI {
		return false
	}
	info, err := os.Stdin.Stat()
	return err == nil && (info.Mode()&os.ModeCharDevice) != 0 && term.IsTerminal(int(os.Stdin.Fd()))
}

func runInitIntegrationWizard(portableConfig *harness.PortableConfig, integrations *harness.Integrations) error {
	if portableConfig == nil || integrations == nil {
		return nil
	}
	reader := bufio.NewReader(os.Stdin)
	probe := harness.RealSystemProbe{}
	engram := harness.DetectEngramWithConfig(probe, portableConfig)
	stitch := harness.DetectStitchWithConfig(probe, portableConfig)
	opendesign := harness.DetectOpenDesignWithConfig(probe, portableConfig)
	openpencil := harness.DetectOpenPencilWithConfig(probe, portableConfig)
	options := initIntegrationOptions(engram, stitch, opendesign, openpencil)

	selected, canceled, err := runInitIntegrationSelector(options)
	if err != nil {
		PrintInfo(fmt.Sprintf("selector interactivo no disponible (%s); usando fallback textual", err))
		clearInitScreen()
		fmt.Println("Shipwright Init — Integraciones")
		fmt.Println("================================")
		renderInitIntegrationChecklist(options, nil)
		selected = askInitIntegrationSelection(reader, options)
	}
	if canceled {
		PrintInfo("Wizard de integraciones omitido; se conserva la configuración inicial.")
		return nil
	}

	clearInitScreen()
	fmt.Println("Shipwright Init — Integraciones")
	fmt.Println("================================")
	fmt.Println("Selección confirmada:")
	fmt.Println()
	renderInitIntegrationChecklist(options, selected)

	if selected["engram"] {
		integrations.EnableEngram()
		PrintSuccess("Engram activado")
	} else {
		integrations.DisableEngram()
		PrintInfo("Engram desactivado; se usará el log local")
	}

	if selected["stitch"] {
		integrations.EnableStitch()
		stitch = configureStitchCredentials(reader, portableConfig, stitch)
		integrations.ApplyStitchDetection(stitch)
		if stitch.Available {
			PrintSuccess("Stitch configurado localmente")
			validateStitchDuringInit(portableConfig)
		} else {
			PrintInfo("Stitch quedó activado, pero todavía faltan credenciales")
			PrintInfo("Seteá STITCH_API_KEY o guardalo luego con el wizard/secret local")
		}
	} else {
		integrations.DisableStitch()
		PrintInfo("Stitch desactivado; diseño usará OpenPencil si está activo o doc-only")
	}

	if selected["opendesign"] {
		integrations.EnableOpenDesign()
		opendesign = configureOpenDesignDuringInit(reader, portableConfig, opendesign)
		integrations.ApplyOpenDesignDetection(opendesign)
		if opendesign.Available {
			PrintSuccess("OpenDesign configurado localmente")
		} else {
			PrintInfo("OpenDesign quedó activado, pero el daemon/socket no está verificado todavía")
		}
	} else {
		integrations.DisableOpenDesign()
		PrintInfo("OpenDesign desactivado")
	}

	if selected["openpencil"] {
		integrations.EnableOpenPencil()
		PrintSuccess("OpenPencil activado como integración opcional")
		if !openpencil.Active {
			PrintInfo("Shipwright CLI no puede verificar el canvas vivo; OpenCode debe ver tools open-pencil_*.")
		}
	} else {
		integrations.DisableOpenPencil()
		PrintInfo("OpenPencil desactivado")
	}

	integrations.ApplyDetection(engram, openpencil)
	integrations.ApplyOpenDesignDetection(opendesign)
	integrations.ApplyPortableConfig(portableConfig)
	if err := portableConfig.Save(); err != nil {
		return err
	}
	return integrations.Save()
}

func clearInitScreen() {
	fmt.Print("\033[H\033[2J")
}

func initIntegrationOptions(engram, stitch, opendesign, openpencil harness.DetectionResult) []initIntegrationOption {
	return []initIntegrationOption{
		{
			Number:      1,
			Key:         "engram",
			Name:        "Engram",
			Description: "memoria persistente y decisiones del proyecto",
			Recommended: engram.Available,
			Detection:   engram,
		},
		{
			Number:      2,
			Key:         "stitch",
			Name:        "Stitch",
			Description: "diseño high-fidelity principal",
			Recommended: true,
			Detection:   stitch,
		},
		{
			Number:      3,
			Key:         "opendesign",
			Name:        "OpenDesign",
			Description: "diseño/canvas MCP local",
			Recommended: opendesign.Available,
			Detection:   opendesign,
		},
		{
			Number:      4,
			Key:         "openpencil",
			Name:        "OpenPencil",
			Description: "canvas local opcional",
			Recommended: false,
			Detection:   openpencil,
		},
	}
}

func runInitIntegrationSelector(options []initIntegrationOption) (map[string]bool, bool, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, false, err
	}
	defer func() {
		_ = term.Restore(fd, oldState)
	}()

	state := newInitIntegrationSelectorState(options)
	for {
		renderInitIntegrationSelector(state)
		key, err := readInitSelectorKey(os.Stdin)
		if err != nil {
			return nil, false, err
		}
		switch key {
		case initSelectorKeyUp:
			state.Move(-1)
		case initSelectorKeyDown:
			state.Move(1)
		case initSelectorKeyToggle:
			state.Toggle()
		case initSelectorKeySubmit:
			return state.CloneSelected(), false, nil
		case initSelectorKeyCancel:
			return nil, true, nil
		}
	}
}

func newInitIntegrationSelectorState(options []initIntegrationOption) *initIntegrationSelectorState {
	selected := map[string]bool{}
	for _, option := range options {
		if option.Recommended {
			selected[option.Key] = true
		}
	}
	return &initIntegrationSelectorState{
		Options:  options,
		Selected: selected,
	}
}

func (s *initIntegrationSelectorState) Move(delta int) {
	if s == nil || len(s.Options) == 0 {
		return
	}
	s.Cursor += delta
	if s.Cursor < 0 {
		s.Cursor = len(s.Options) - 1
	}
	if s.Cursor >= len(s.Options) {
		s.Cursor = 0
	}
}

func (s *initIntegrationSelectorState) Toggle() {
	if s == nil || len(s.Options) == 0 {
		return
	}
	key := s.Options[s.Cursor].Key
	s.Selected[key] = !s.Selected[key]
}

func (s *initIntegrationSelectorState) CloneSelected() map[string]bool {
	clone := map[string]bool{}
	if s == nil {
		return clone
	}
	for key, enabled := range s.Selected {
		if enabled {
			clone[key] = true
		}
	}
	return clone
}

func renderInitIntegrationSelector(state *initIntegrationSelectorState) {
	fmt.Print(renderInitIntegrationSelectorView(state))
}

func renderInitIntegrationSelectorView(state *initIntegrationSelectorState) string {
	var builder strings.Builder
	writeRawLine(&builder, "\033[2J\033[HShipwright Init — Integraciones")
	writeRawLine(&builder, "================================")
	writeRawLine(&builder, "")
	writeRawLine(&builder, "↑/↓ o j/k: moverte · Space: marcar/desmarcar · Enter: continuar · q: omitir")
	writeRawLine(&builder, "")
	for index, option := range state.Options {
		cursor := " "
		if index == state.Cursor {
			cursor = ">"
		}
		checked := " "
		if state.Selected[option.Key] {
			checked = "X"
		}
		status := shortIntegrationStatus(option.Detection)
		title := fmt.Sprintf("%s [%s] %s", cursor, checked, option.Name)
		if status != "" {
			title += " (" + status + ")"
		}
		writeRawLine(&builder, title)
		writeRawLine(&builder, "    "+option.Description)
	}
	return builder.String()
}

func writeRawLine(builder *strings.Builder, line string) {
	builder.WriteString(line)
	builder.WriteString("\r\n")
}

func readInitSelectorKey(input io.Reader) (initSelectorKey, error) {
	var first [1]byte
	if _, err := input.Read(first[:]); err != nil {
		return initSelectorKeyUnknown, err
	}
	switch first[0] {
	case 'k', 'K':
		return initSelectorKeyUp, nil
	case 'j', 'J':
		return initSelectorKeyDown, nil
	case ' ', 'x', 'X':
		return initSelectorKeyToggle, nil
	case '\r', '\n':
		return initSelectorKeySubmit, nil
	case 'q', 'Q':
		return initSelectorKeyCancel, nil
	case 27:
		var rest [2]byte
		if _, err := io.ReadFull(input, rest[:]); err != nil {
			return initSelectorKeyCancel, nil
		}
		if rest[0] == '[' {
			switch rest[1] {
			case 'A':
				return initSelectorKeyUp, nil
			case 'B':
				return initSelectorKeyDown, nil
			}
		}
	}
	return initSelectorKeyUnknown, nil
}

func renderInitIntegrationChecklist(options []initIntegrationOption, selected map[string]bool) {
	for _, option := range options {
		checked := " "
		if selected != nil && selected[option.Key] {
			checked = "X"
		}
		fmt.Printf("[%s] %d. %-10s — %s", checked, option.Number, option.Name, option.Description)
		status := shortIntegrationStatus(option.Detection)
		if status != "" {
			fmt.Printf(" (%s)", status)
		}
		fmt.Println()
	}
	fmt.Println()
}

func askInitIntegrationSelection(reader *bufio.Reader, options []initIntegrationOption) map[string]bool {
	for {
		fmt.Print("Seleccioná integraciones (ej: 1,2 o engram,stitch; Enter=recomendado; 0=ninguna): ")
		answer, _ := reader.ReadString('\n')
		selected, err := parseInitIntegrationSelection(answer, options)
		if err == nil {
			return selected
		}
		fmt.Println(err.Error())
	}
}

func parseInitIntegrationSelection(input string, options []initIntegrationOption) (map[string]bool, error) {
	selected := map[string]bool{}
	byNumber := map[int]initIntegrationOption{}
	byKey := map[string]initIntegrationOption{}
	for _, option := range options {
		byNumber[option.Number] = option
		byKey[strings.ToLower(option.Key)] = option
		byKey[strings.ToLower(option.Name)] = option
		if option.Recommended {
			selected[option.Key] = true
		}
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return selected, nil
	}
	if input == "0" || input == "none" || input == "ninguna" || input == "ninguno" {
		return map[string]bool{}, nil
	}

	selected = map[string]bool{}
	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == ';' || r == ' '
	})
	for _, raw := range parts {
		part := strings.TrimSpace(raw)
		if part == "" {
			continue
		}
		if number, err := strconv.Atoi(part); err == nil {
			option, ok := byNumber[number]
			if !ok {
				return nil, fmt.Errorf("Integración desconocida: %d", number)
			}
			selected[option.Key] = true
			continue
		}
		option, ok := byKey[part]
		if !ok {
			return nil, fmt.Errorf("Integración desconocida: %s", part)
		}
		selected[option.Key] = true
	}
	return selected, nil
}

func shortIntegrationStatus(result harness.DetectionResult) string {
	switch result.Status {
	case harness.DetectionAvailable:
		return "disponible"
	case harness.DetectionInstalledNoCanvas:
		return "instalado, canvas no verificado"
	case harness.DetectionConfiguredUnverified:
		return "configurado sin verificar"
	case harness.DetectionNotInstalled:
		return "no configurado"
	default:
		return result.Status
	}
}

func askInitYesNo(reader *bufio.Reader, question string, defaultYes bool) bool {
	suffix := " [Y/n]: "
	if !defaultYes {
		suffix = " [y/N]: "
	}
	for {
		fmt.Print(question + suffix)
		answer, _ := reader.ReadString('\n')
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer == "" {
			return defaultYes
		}
		switch answer {
		case "y", "yes", "s", "si", "sí":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("Respondé sí o no.")
		}
	}
}

func askInitValue(reader *bufio.Reader, question string) string {
	fmt.Print(question + ": ")
	answer, _ := reader.ReadString('\n')
	return strings.TrimSpace(answer)
}

func configureStitchCredentials(reader *bufio.Reader, portableConfig *harness.PortableConfig, current harness.DetectionResult) harness.DetectionResult {
	if current.Available {
		PrintInfo("Stitch ya tiene credenciales disponibles.")
		return current
	}

	apiKeyEnv := portableConfig.Integrations.Stitch.APIKeyEnv
	if apiKeyEnv == "" {
		apiKeyEnv = "STITCH_API_KEY"
	}
	fmt.Println()
	fmt.Println("Credenciales Stitch")
	fmt.Println("Pegá tu API key de Stitch si querés validarlo ahora.")
	fmt.Println("Ojo: por compatibilidad multiplataforma, la entrada se ve en la terminal.")
	fmt.Println("Se guardará en .harness/secrets.local.env y esa ruta queda ignorada por .harness/.gitignore.")
	apiKey := askInitValue(reader, apiKeyEnv+" (Enter para omitir)")
	if apiKey == "" {
		return harness.DetectStitchWithConfig(harness.RealSystemProbe{}, portableConfig)
	}
	if err := harness.SaveLocalSecret(apiKeyEnv, apiKey); err != nil {
		PrintInfo(fmt.Sprintf("warning: no se pudo guardar secret local: %s", err))
	} else {
		PrintSuccess(fmt.Sprintf("%s guardado como secret local", apiKeyEnv))
	}
	_ = os.Setenv(apiKeyEnv, apiKey)
	return harness.DetectStitchWithConfig(harness.RealSystemProbe{}, portableConfig)
}

func validateStitchDuringInit(portableConfig *harness.PortableConfig) {
	apiKeyEnv := portableConfig.Integrations.Stitch.APIKeyEnv
	if apiKeyEnv == "" {
		apiKeyEnv = "STITCH_API_KEY"
	}
	apiKey := strings.TrimSpace(os.Getenv(apiKeyEnv))
	if apiKey == "" {
		apiKey = harness.LocalSecretValue(apiKeyEnv)
	}
	if apiKey == "" {
		PrintInfo("Stitch: validación de conexión omitida porque no hay API key")
		return
	}
	endpoint := portableConfig.Integrations.Stitch.MCPURL
	if endpoint == "" {
		endpoint = harness.DefaultStitchMCPEndpoint
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	result := harness.ValidateStitchConnection(ctx, &http.Client{Timeout: 10 * time.Second}, endpoint, apiKey)
	if result.Healthy {
		PrintSuccess("Stitch connection check OK")
		return
	}
	if result.Status == "skipped" {
		PrintInfo("Stitch connection check skipped: " + result.Detail)
		return
	}
	PrintInfo(fmt.Sprintf("Stitch connection check no quedó saludable (%s). Revisá la API key o el endpoint.", result.Status))
	if result.Detail != "" {
		PrintInfo("detalle: " + result.Detail)
	}
}

func configureOpenDesignDuringInit(reader *bufio.Reader, portableConfig *harness.PortableConfig, current harness.DetectionResult) harness.DetectionResult {
	probe := harness.RealSystemProbe{}
	if current.Configured && strings.TrimSpace(portableConfig.Integrations.OpenDesign.MCPCommand) != "" {
		PrintInfo("OpenDesign ya tiene comando MCP configurado.")
		return harness.DetectOpenDesignWithConfig(probe, portableConfig)
	}

	fmt.Println()
	fmt.Println("Configuración OpenDesign")
	fmt.Println("Buscando configuración local de OpenDesign...")
	autoDetected := harness.AutoConfigureOpenDesign(probe, portableConfig)
	if autoDetected.Configured && strings.TrimSpace(portableConfig.Integrations.OpenDesign.MCPCommand) != "" {
		PrintSuccess("OpenDesign autodetectado")
		fmt.Printf("  command: %s\n", portableConfig.Integrations.OpenDesign.MCPCommand)
		if len(portableConfig.Integrations.OpenDesign.MCPArgs) > 0 {
			fmt.Printf("  args:    %s\n", strings.Join(portableConfig.Integrations.OpenDesign.MCPArgs, " "))
		}
		if portableConfig.Integrations.OpenDesign.DataDir != "" {
			fmt.Printf("  data:    %s\n", portableConfig.Integrations.OpenDesign.DataDir)
		}
		if portableConfig.Integrations.OpenDesign.IPCPath != "" {
			fmt.Printf("  ipc:     %s\n", portableConfig.Integrations.OpenDesign.IPCPath)
		}
		if !autoDetected.Available && autoDetected.Reason != "" {
			PrintInfo(autoDetected.Reason)
		}
		return autoDetected
	}

	PrintInfo("No pude autodetectar OpenDesign. Paso a configuración manual.")
	fmt.Println("Tip: si OpenDesign tiene el CLI `od` instalado, asegurate de que esté en PATH y reintentá `shipwright init`.")
	fmt.Println("Si usás el repo local de OpenDesign, podés setear OPENDESIGN_ROOT y Shipwright buscará apps/daemon/dist/cli.js.")
	fmt.Println()
	fmt.Println("Configuración manual OpenDesign")
	fmt.Println("Dejá vacío cualquier campo que no conozcas; Shipwright usará fallback doc-only si no alcanza.")
	command := askInitValue(reader, "Comando MCP o Node (ej: od, /opt/homebrew/bin/node; Enter para omitir)")
	if command == "" {
		return harness.DetectOpenDesignWithConfig(probe, portableConfig)
	}
	cliArg := askInitValue(reader, "Path a cli.js si usás Node (Enter si el comando ya es od)")
	modeArg := askInitValue(reader, "Argumento MCP (Enter=mcp)")
	if modeArg == "" {
		modeArg = "mcp"
	}
	dataDir := askInitValue(reader, "Carpeta de datos OD_DATA_DIR (Enter para omitir)")
	ipcPath := askInitValue(reader, "Socket OD_SIDECAR_IPC_PATH (Enter para usar default si aplica)")
	if ipcPath == "" && cliArg != "" {
		ipcPath = "/tmp/open-design/ipc/default/daemon.sock"
	}

	portableConfig.Integrations.OpenDesign.MCPCommand = command
	portableConfig.Integrations.OpenDesign.MCPArgs = []string{}
	if cliArg != "" {
		portableConfig.Integrations.OpenDesign.MCPArgs = append(portableConfig.Integrations.OpenDesign.MCPArgs, cliArg)
	}
	portableConfig.Integrations.OpenDesign.MCPArgs = append(portableConfig.Integrations.OpenDesign.MCPArgs, modeArg)
	portableConfig.Integrations.OpenDesign.DataDir = dataDir
	portableConfig.Integrations.OpenDesign.IPCPath = ipcPath
	portableConfig.Integrations.OpenDesign.Mode = harness.ConfigModeMCP
	portableConfig.Integrations.OpenDesign.Fallback = "design-doc-only"
	return harness.DetectOpenDesignWithConfig(probe, portableConfig)
}
