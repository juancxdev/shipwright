package harness

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	DetectionPathBinary    = "binary"
	DetectionPathApp       = "app"
	DetectionPathMCPServer = "mcp_server"

	DetectionNotInstalled         = "not_installed"
	DetectionInstalled            = "installed"
	DetectionAvailable            = "available"
	DetectionConfiguredUnverified = "configured_unverified"
	DetectionInstalledNoCanvas    = "installed_no_active_canvas"
	DetectionUnavailableFallback  = "unavailable_fallback"
)

type DetectionResult struct {
	Name       string       `json:"name"`
	Platform   PlatformInfo `json:"platform"`
	Installed  bool         `json:"installed"`
	Configured bool         `json:"configured"`
	Available  bool         `json:"available"`
	Active     bool         `json:"active"`
	Version    string       `json:"version,omitempty"`
	Path       string       `json:"path,omitempty"`
	PathKind   string       `json:"path_kind,omitempty"`
	Status     string       `json:"status"`
	Reason     string       `json:"reason,omitempty"`
	Fallback   string       `json:"fallback,omitempty"`
}

func DetectEngram(probe SystemProbe) DetectionResult {
	return DetectEngramWithConfig(probe, nil)
}

func DetectEngramWithConfig(probe SystemProbe, cfg *PortableConfig) DetectionResult {
	if probe == nil {
		probe = RealSystemProbe{}
	}
	platform := DetectPlatform(probe)
	fallback := DecisionsFile
	if cfg != nil && cfg.Integrations.Engram.Fallback != "" {
		fallback = cfg.Integrations.Engram.Fallback
	}
	binaryName := "engram"
	if platform.OS == "windows" {
		binaryName = "engram.exe"
	}
	if cfg != nil && strings.TrimSpace(cfg.Integrations.Engram.BinaryPath) != "" {
		return detectConfiguredBinary(probe, platform, "engram", strings.TrimSpace(cfg.Integrations.Engram.BinaryPath), fallback)
	}
	if override := strings.TrimSpace(probe.Getenv("ENGRAM_BINARY")); override != "" {
		return detectConfiguredBinary(probe, platform, "engram", override, fallback)
	}
	path, err := probe.LookPath(binaryName)
	if err != nil && platform.OS == "windows" {
		path, err = probe.LookPath("engram")
	}
	if err != nil {
		return DetectionResult{
			Name:     "engram",
			Platform: platform,
			Status:   DetectionNotInstalled,
			Reason:   fmt.Sprintf("%s not found in PATH and ENGRAM_BINARY not set", binaryName),
			Fallback: fallback,
		}
	}
	return DetectionResult{
		Name:       "engram",
		Platform:   platform,
		Installed:  true,
		Configured: true,
		Available:  true,
		Path:       path,
		PathKind:   DetectionPathBinary,
		Status:     DetectionAvailable,
		Reason:     "engram binary found in PATH",
		Fallback:   fallback,
	}
}

func detectConfiguredBinary(probe SystemProbe, platform PlatformInfo, name, path, fallback string) DetectionResult {
	info, err := probe.Stat(path)
	if err != nil {
		return DetectionResult{
			Name:       name,
			Platform:   platform,
			Configured: true,
			Status:     DetectionNotInstalled,
			Path:       path,
			PathKind:   DetectionPathBinary,
			Reason:     fmt.Sprintf("configured path not found: %s", path),
			Fallback:   fallback,
		}
	}
	if info.IsDir() {
		return DetectionResult{
			Name:       name,
			Platform:   platform,
			Configured: true,
			Status:     DetectionNotInstalled,
			Path:       path,
			PathKind:   DetectionPathBinary,
			Reason:     "configured binary path is a directory",
			Fallback:   fallback,
		}
	}
	return DetectionResult{
		Name:       name,
		Platform:   platform,
		Installed:  true,
		Configured: true,
		Available:  true,
		Path:       path,
		PathKind:   DetectionPathBinary,
		Status:     DetectionAvailable,
		Reason:     "configured binary path exists",
		Fallback:   fallback,
	}
}

func DetectOpenPencil(probe SystemProbe) DetectionResult {
	return DetectOpenPencilWithConfig(probe, nil)
}

func DetectOpenPencilWithConfig(probe SystemProbe, cfg *PortableConfig) DetectionResult {
	if probe == nil {
		probe = RealSystemProbe{}
	}
	platform := DetectPlatform(probe)
	fallback := "design-doc-only"
	if cfg != nil && cfg.Integrations.OpenPencil.Fallback != "" {
		fallback = cfg.Integrations.OpenPencil.Fallback
	}
	configuredMCP := strings.TrimSpace(probe.Getenv("OPENPENCIL_MCP_SERVER"))
	if cfg != nil && strings.TrimSpace(cfg.Integrations.OpenPencil.MCPServerPath) != "" {
		configuredMCP = strings.TrimSpace(cfg.Integrations.OpenPencil.MCPServerPath)
	}
	if configuredMCP != "" {
		return detectOpenPencilMCPPath(probe, platform, configuredMCP, true, fallback)
	}

	configuredCommand := strings.TrimSpace(probe.Getenv("OPENPENCIL_MCP_COMMAND"))
	if cfg != nil && strings.TrimSpace(cfg.Integrations.OpenPencil.MCPCommand) != "" {
		configuredCommand = strings.TrimSpace(cfg.Integrations.OpenPencil.MCPCommand)
	}
	if configuredCommand != "" {
		return detectOpenPencilMCPCommand(probe, platform, configuredCommand, true, fallback)
	}
	if path, err := probe.LookPath("openpencil-mcp"); err == nil && strings.TrimSpace(path) != "" {
		return detectOpenPencilMCPCommand(probe, platform, path, false, fallback)
	}

	candidates := openPencilMCPServerCandidates(probe, platform, cfg)
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if info, err := probe.Stat(candidate); err == nil && !info.IsDir() {
			return detectOpenPencilMCPPath(probe, platform, candidate, false, fallback)
		}
	}

	appPath := strings.TrimSpace(probe.Getenv("OPENPENCIL_APP_PATH"))
	if cfg != nil && strings.TrimSpace(cfg.Integrations.OpenPencil.AppPath) != "" {
		appPath = strings.TrimSpace(cfg.Integrations.OpenPencil.AppPath)
	}
	if appPath != "" {
		if info, err := probe.Stat(appPath); err == nil && info.IsDir() {
			return DetectionResult{
				Name:       "openpencil",
				Platform:   platform,
				Installed:  true,
				Configured: true,
				Status:     DetectionConfiguredUnverified,
				Path:       appPath,
				PathKind:   DetectionPathApp,
				Reason:     "OPENPENCIL_APP_PATH exists but MCP server was not found; set OPENPENCIL_MCP_SERVER",
				Fallback:   fallback,
			}
		}
		return DetectionResult{
			Name:       "openpencil",
			Platform:   platform,
			Configured: true,
			Status:     DetectionNotInstalled,
			Path:       appPath,
			PathKind:   DetectionPathApp,
			Reason:     "OPENPENCIL_APP_PATH not found",
			Fallback:   fallback,
		}
	}

	return DetectionResult{
		Name:     "openpencil",
		Platform: platform,
		Status:   DetectionNotInstalled,
		Reason:   "OpenPencil MCP server not found; set OPENPENCIL_MCP_SERVER or OPENPENCIL_APP_PATH",
		Fallback: fallback,
	}
}

func detectOpenPencilMCPCommand(probe SystemProbe, platform PlatformInfo, command string, configured bool, fallback string) DetectionResult {
	resolved := command
	if path, err := probe.LookPath(command); err == nil && strings.TrimSpace(path) != "" {
		resolved = path
	} else if configured {
		return DetectionResult{
			Name:       "openpencil",
			Platform:   platform,
			Configured: configured,
			Status:     DetectionNotInstalled,
			Path:       command,
			PathKind:   DetectionPathBinary,
			Reason:     "OpenPencil MCP command not found in PATH",
			Fallback:   fallback,
		}
	}

	canvasActive := strings.EqualFold(strings.TrimSpace(probe.Getenv("OPENPENCIL_CANVAS_ACTIVE")), "true")
	status := DetectionInstalledNoCanvas
	available := false
	active := false
	reason := "OpenPencil MCP command found; canvas not verified by Shipwright CLI. Validate through the OpenCode MCP tool call."
	if canvasActive {
		status = DetectionAvailable
		available = true
		active = true
		reason = "OpenPencil MCP command found and canvas reported active"
	}
	return DetectionResult{
		Name:       "openpencil",
		Platform:   platform,
		Installed:  true,
		Configured: configured,
		Available:  available,
		Active:     active,
		Path:       resolved,
		PathKind:   DetectionPathBinary,
		Status:     status,
		Reason:     reason,
		Fallback:   fallback,
	}
}

func detectOpenPencilMCPPath(probe SystemProbe, platform PlatformInfo, path string, configured bool, fallback string) DetectionResult {
	info, err := probe.Stat(path)
	if err != nil {
		return DetectionResult{
			Name:       "openpencil",
			Platform:   platform,
			Configured: configured,
			Status:     DetectionNotInstalled,
			Path:       path,
			PathKind:   DetectionPathMCPServer,
			Reason:     "OpenPencil MCP server path not found",
			Fallback:   fallback,
		}
	}
	if info.IsDir() {
		return DetectionResult{
			Name:       "openpencil",
			Platform:   platform,
			Configured: configured,
			Status:     DetectionNotInstalled,
			Path:       path,
			PathKind:   DetectionPathMCPServer,
			Reason:     "OpenPencil MCP server path is a directory",
			Fallback:   fallback,
		}
	}
	canvasActive := strings.EqualFold(strings.TrimSpace(probe.Getenv("OPENPENCIL_CANVAS_ACTIVE")), "true")
	status := DetectionInstalledNoCanvas
	available := false
	active := false
	reason := "OpenPencil MCP server found; canvas not verified by Shipwright CLI. Validate through the OpenCode MCP tool call."
	if canvasActive {
		status = DetectionAvailable
		available = true
		active = true
		reason = "OpenPencil MCP server found and canvas reported active"
	}
	return DetectionResult{
		Name:       "openpencil",
		Platform:   platform,
		Installed:  true,
		Configured: configured,
		Available:  available,
		Active:     active,
		Path:       path,
		PathKind:   DetectionPathMCPServer,
		Status:     status,
		Reason:     reason,
		Fallback:   fallback,
	}
}

func openPencilMCPServerCandidates(probe SystemProbe, platform PlatformInfo, cfg *PortableConfig) []string {
	home := platform.HomeDir
	appPath := strings.TrimSpace(probe.Getenv("OPENPENCIL_APP_PATH"))
	if cfg != nil && strings.TrimSpace(cfg.Integrations.OpenPencil.AppPath) != "" {
		appPath = strings.TrimSpace(cfg.Integrations.OpenPencil.AppPath)
	}
	var candidates []string
	if appPath != "" {
		candidates = append(candidates, filepath.Join(appPath, "Contents", "Resources", "mcp-server.cjs"))
	}
	switch platform.OS {
	case "darwin":
		candidates = append(candidates,
			"/Applications/OpenPencil.app/Contents/Resources/mcp-server.cjs",
			filepath.Join(home, "Applications", "OpenPencil.app", "Contents", "Resources", "mcp-server.cjs"),
		)
	case "windows":
		localAppData := strings.TrimSpace(probe.Getenv("LOCALAPPDATA"))
		programFiles := strings.TrimSpace(probe.Getenv("PROGRAMFILES"))
		candidates = append(candidates,
			filepath.Join(localAppData, "OpenPencil", "resources", "mcp-server.cjs"),
			filepath.Join(programFiles, "OpenPencil", "resources", "mcp-server.cjs"),
		)
	case "linux":
		candidates = append(candidates,
			filepath.Join(home, ".local", "share", "OpenPencil", "resources", "mcp-server.cjs"),
			"/opt/OpenPencil/resources/mcp-server.cjs",
		)
	}
	return candidates
}

func DetectStitch(probe SystemProbe) DetectionResult {
	return DetectStitchWithConfig(probe, nil)
}

func DetectStitchWithConfig(probe SystemProbe, cfg *PortableConfig) DetectionResult {
	if probe == nil {
		probe = RealSystemProbe{}
	}
	platform := DetectPlatform(probe)
	fallback := "design-doc-only"
	apiKeyEnv := "STITCH_API_KEY"
	accessTokenEnv := "STITCH_ACCESS_TOKEN"
	projectEnv := "GOOGLE_CLOUD_PROJECT"
	if cfg != nil {
		if strings.TrimSpace(cfg.Integrations.Stitch.Fallback) != "" {
			fallback = cfg.Integrations.Stitch.Fallback
		}
		if strings.TrimSpace(cfg.Integrations.Stitch.APIKeyEnv) != "" {
			apiKeyEnv = strings.TrimSpace(cfg.Integrations.Stitch.APIKeyEnv)
		}
		if strings.TrimSpace(cfg.Integrations.Stitch.AccessTokenEnv) != "" {
			accessTokenEnv = strings.TrimSpace(cfg.Integrations.Stitch.AccessTokenEnv)
		}
		if strings.TrimSpace(cfg.Integrations.Stitch.ProjectEnv) != "" {
			projectEnv = strings.TrimSpace(cfg.Integrations.Stitch.ProjectEnv)
		}
	}

	apiKey := strings.TrimSpace(probe.Getenv(apiKeyEnv))
	accessToken := strings.TrimSpace(probe.Getenv(accessTokenEnv))
	project := strings.TrimSpace(probe.Getenv(projectEnv))
	if apiKey == "" {
		apiKey = LocalSecretValue(apiKeyEnv)
	}
	if accessToken == "" {
		accessToken = LocalSecretValue(accessTokenEnv)
	}
	if project == "" {
		project = LocalSecretValue(projectEnv)
	}
	if apiKey != "" || (accessToken != "" && project != "") {
		return DetectionResult{
			Name:       "stitch",
			Platform:   platform,
			Installed:  true,
			Configured: true,
			Available:  true,
			Active:     true,
			Status:     DetectionAvailable,
			Reason:     "Stitch credentials found in environment or .harness/secrets.local.env",
			Fallback:   fallback,
		}
	}

	if accessToken != "" && project == "" {
		return DetectionResult{
			Name:       "stitch",
			Platform:   platform,
			Installed:  true,
			Configured: true,
			Status:     DetectionConfiguredUnverified,
			Reason:     accessTokenEnv + " is set but " + projectEnv + " is missing",
			Fallback:   fallback,
		}
	}

	return DetectionResult{
		Name:     "stitch",
		Platform: platform,
		Status:   DetectionNotInstalled,
		Reason:   "Stitch credentials not found; set STITCH_API_KEY or STITCH_ACCESS_TOKEN + GOOGLE_CLOUD_PROJECT",
		Fallback: fallback,
	}
}

func DetectOpenDesign(probe SystemProbe) DetectionResult {
	return DetectOpenDesignWithConfig(probe, nil)
}

func DetectOpenDesignWithConfig(probe SystemProbe, cfg *PortableConfig) DetectionResult {
	if probe == nil {
		probe = RealSystemProbe{}
	}
	platform := DetectPlatform(probe)
	fallback := "design-doc-only"
	command := strings.TrimSpace(probe.Getenv("OPENDESIGN_MCP_COMMAND"))
	dataDir := firstNonEmpty(
		strings.TrimSpace(probe.Getenv("OPENDESIGN_DATA_DIR")),
		strings.TrimSpace(probe.Getenv("OD_DATA_DIR")),
	)
	ipcPath := firstNonEmpty(
		strings.TrimSpace(probe.Getenv("OPENDESIGN_SIDECAR_IPC_PATH")),
		strings.TrimSpace(probe.Getenv("OD_SIDECAR_IPC_PATH")),
	)
	if cfg != nil {
		if strings.TrimSpace(cfg.Integrations.OpenDesign.Fallback) != "" {
			fallback = cfg.Integrations.OpenDesign.Fallback
		}
		if strings.TrimSpace(cfg.Integrations.OpenDesign.MCPCommand) != "" {
			command = strings.TrimSpace(cfg.Integrations.OpenDesign.MCPCommand)
		}
		if strings.TrimSpace(cfg.Integrations.OpenDesign.DataDir) != "" {
			dataDir = strings.TrimSpace(cfg.Integrations.OpenDesign.DataDir)
		}
		if strings.TrimSpace(cfg.Integrations.OpenDesign.IPCPath) != "" {
			ipcPath = strings.TrimSpace(cfg.Integrations.OpenDesign.IPCPath)
		}
	}
	if command == "" {
		return DetectionResult{
			Name:     "opendesign",
			Platform: platform,
			Status:   DetectionNotInstalled,
			Reason:   "OpenDesign MCP command not configured; run 'shipwright integrations configure opendesign'.",
			Fallback: fallback,
		}
	}

	resolved := command
	if strings.ContainsAny(command, `/\\`) {
		info, err := probe.Stat(command)
		if err != nil {
			return DetectionResult{
				Name:       "opendesign",
				Platform:   platform,
				Configured: true,
				Status:     DetectionNotInstalled,
				Path:       command,
				PathKind:   DetectionPathBinary,
				Reason:     "OpenDesign MCP command path not found",
				Fallback:   fallback,
			}
		}
		if info.IsDir() {
			return DetectionResult{
				Name:       "opendesign",
				Platform:   platform,
				Configured: true,
				Status:     DetectionNotInstalled,
				Path:       command,
				PathKind:   DetectionPathBinary,
				Reason:     "OpenDesign MCP command path is a directory",
				Fallback:   fallback,
			}
		}
	} else if path, err := probe.LookPath(command); err == nil && strings.TrimSpace(path) != "" {
		resolved = path
	} else {
		return DetectionResult{
			Name:       "opendesign",
			Platform:   platform,
			Configured: true,
			Status:     DetectionNotInstalled,
			Path:       command,
			PathKind:   DetectionPathBinary,
			Reason:     "OpenDesign MCP command not found in PATH",
			Fallback:   fallback,
		}
	}

	status := DetectionConfiguredUnverified
	available := false
	active := false
	reason := "OpenDesign MCP command configured; sidecar IPC not verified. Start OpenDesign daemon/app, then verify with OpenCode MCP tools."
	if ipcPath != "" {
		if info, err := probe.Stat(ipcPath); err == nil && !info.IsDir() {
			status = DetectionAvailable
			available = true
			active = true
			reason = "OpenDesign MCP command and sidecar IPC path are available"
		} else {
			reason = "OpenDesign MCP command configured, but sidecar IPC path is not available yet"
		}
	}
	if dataDir == "" {
		reason += "; OD_DATA_DIR is not configured"
	}

	return DetectionResult{
		Name:       "opendesign",
		Platform:   platform,
		Installed:  true,
		Configured: true,
		Available:  available,
		Active:     active,
		Path:       resolved,
		PathKind:   DetectionPathBinary,
		Status:     status,
		Reason:     reason,
		Fallback:   fallback,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
