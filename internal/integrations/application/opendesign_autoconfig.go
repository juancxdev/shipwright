package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	config "shipwright/internal/config/application"
	platform "shipwright/internal/platform/application"
)

const openDesignInstallInfoPath = "/api/mcp/install-info"

type openDesignInstallInfo struct {
	Command    string            `json:"command"`
	Args       []string          `json:"args"`
	Env        map[string]string `json:"env"`
	DaemonURL  string            `json:"daemonUrl"`
	WebBaseURL string            `json:"webBaseUrl"`
	CliExists  bool              `json:"cliExists"`
	NodeExists bool              `json:"nodeExists"`
	BuildHint  string            `json:"buildHint"`
}

var fetchOpenDesignInstallInfo = fetchOpenDesignInstallInfoHTTP

// AutoConfigureOpenDesign attempts to fill the portable OpenDesign MCP config from
// an existing explicit config, the running OpenDesign daemon install-info API,
// installed CLIs, or common local source checkouts.
// It mutates cfg when a usable command can be inferred.
func AutoConfigureOpenDesign(probe platform.SystemProbe, cfg *config.PortableConfig) DetectionResult {
	if probe == nil {
		probe = platform.RealSystemProbe{}
	}
	if cfg == nil {
		cfg = config.DefaultPortableConfig()
	}
	cfg.ApplyEnv(probe)
	cfg.Normalize()

	if normalizeOpenDesignExistingConfig(probe, cfg) {
		return DetectOpenDesignWithConfig(probe, cfg)
	}

	if info := discoverOpenDesignInstallInfo(probe); info != nil {
		applyOpenDesignInstallInfoConfig(cfg, info)
		return DetectOpenDesignWithConfig(probe, cfg)
	}

	if path, err := probe.LookPath("od"); err == nil && strings.TrimSpace(path) != "" {
		applyOpenDesignConfig(cfg, path, []string{"mcp"}, "", "", defaultOpenDesignIPCPath(probe))
		return DetectOpenDesignWithConfig(probe, cfg)
	}

	if path, err := probe.LookPath("open-design"); err == nil && strings.TrimSpace(path) != "" {
		applyOpenDesignConfig(cfg, path, []string{"mcp"}, "", "", defaultOpenDesignIPCPath(probe))
		return DetectOpenDesignWithConfig(probe, cfg)
	}

	cliPath, dataDir := findOpenDesignDaemonCLI(probe)
	if cliPath == "" {
		return DetectOpenDesignWithConfig(probe, cfg)
	}
	nodePath := findOpenDesignNode(probe)
	if nodePath == "" {
		return DetectOpenDesignWithConfig(probe, cfg)
	}
	applyOpenDesignConfig(cfg, nodePath, []string{cliPath, "mcp"}, "", dataDir, defaultOpenDesignIPCPath(probe))
	return DetectOpenDesignWithConfig(probe, cfg)
}

func discoverOpenDesignInstallInfo(probe platform.SystemProbe) *openDesignInstallInfo {
	for _, baseURL := range openDesignDaemonURLCandidates(probe) {
		info, err := fetchOpenDesignInstallInfo(baseURL)
		if err != nil || info == nil {
			continue
		}
		if !info.usable() {
			continue
		}
		return info
	}
	return nil
}

func (i *openDesignInstallInfo) usable() bool {
	if i == nil {
		return false
	}
	return strings.TrimSpace(i.Command) != "" &&
		len(i.Args) > 0 &&
		i.CliExists &&
		i.NodeExists
}

func openDesignDaemonURLCandidates(probe platform.SystemProbe) []string {
	var values []string
	for _, key := range []string{
		"OPENDESIGN_DAEMON_URL",
		"OPEN_DESIGN_DAEMON_URL",
		"OD_DAEMON_URL",
	} {
		if value := strings.TrimSpace(probe.Getenv(key)); value != "" {
			values = append(values, value)
		}
	}
	values = append(values, discoverOpenDesignToolsDevDaemonURLs(probe)...)
	values = append(values, "http://127.0.0.1:7456")

	seen := map[string]bool{}
	unique := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimRight(strings.TrimSpace(value), "/")
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}

func discoverOpenDesignToolsDevDaemonURLs(probe platform.SystemProbe) []string {
	urls := []string{}
	for _, root := range openDesignRootCandidates(probe) {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		if info, err := probe.Stat(filepath.Join(root, "package.json")); err != nil || info.IsDir() {
			continue
		}
		url, err := openDesignToolsDevDaemonURL(root)
		if err != nil || strings.TrimSpace(url) == "" {
			continue
		}
		urls = append(urls, url)
	}
	return urls
}

func openDesignToolsDevDaemonURL(root string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pnpm", "--silent", "exec", "tools-dev", "status", "--json")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var payload struct {
		Apps struct {
			Daemon struct {
				URL string `json:"url"`
			} `json:"daemon"`
		} `json:"apps"`
		URL string `json:"url"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		return "", err
	}
	return firstNonEmpty(payload.Apps.Daemon.URL, payload.URL), nil
}

func fetchOpenDesignInstallInfoHTTP(baseURL string) (*openDesignInstallInfo, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, errors.New("empty OpenDesign daemon URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+openDesignInstallInfoPath, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("OpenDesign install-info returned HTTP %d", resp.StatusCode)
	}
	var info openDesignInstallInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

func applyOpenDesignInstallInfoConfig(cfg *config.PortableConfig, info *openDesignInstallInfo) {
	dataDir := ""
	ipcPath := ""
	if info != nil && info.Env != nil {
		dataDir = firstNonEmpty(info.Env["OD_DATA_DIR"], info.Env["OPENDESIGN_DATA_DIR"])
		ipcPath = firstNonEmpty(info.Env["OD_SIDECAR_IPC_PATH"], info.Env["OPENDESIGN_SIDECAR_IPC_PATH"])
	}
	applyOpenDesignConfig(cfg, info.Command, info.Args, strings.TrimSpace(info.DaemonURL), dataDir, ipcPath)
}

func normalizeOpenDesignExistingConfig(probe platform.SystemProbe, cfg *config.PortableConfig) bool {
	command := strings.TrimSpace(cfg.Integrations.OpenDesign.MCPCommand)
	if command == "" {
		return false
	}
	if len(cfg.Integrations.OpenDesign.MCPArgs) == 0 && isOpenDesignCLICommand(command) {
		cfg.Integrations.OpenDesign.MCPArgs = []string{"mcp"}
	}
	cfg.Integrations.OpenDesign.Mode = config.ConfigModeMCP
	if cfg.Integrations.OpenDesign.Fallback == "" {
		cfg.Integrations.OpenDesign.Fallback = "design-doc-only"
	}
	if cfg.Integrations.OpenDesign.IPCPath == "" {
		cfg.Integrations.OpenDesign.IPCPath = defaultOpenDesignIPCPath(probe)
	}
	return true
}

func applyOpenDesignConfig(cfg *config.PortableConfig, command string, args []string, daemonURL string, dataDir string, ipcPath string) {
	cfg.Integrations.OpenDesign.MCPCommand = strings.TrimSpace(command)
	cfg.Integrations.OpenDesign.MCPArgs = append([]string{}, args...)
	if strings.TrimSpace(daemonURL) != "" {
		cfg.Integrations.OpenDesign.DaemonURL = strings.TrimRight(strings.TrimSpace(daemonURL), "/")
	}
	if strings.TrimSpace(dataDir) != "" {
		cfg.Integrations.OpenDesign.DataDir = strings.TrimSpace(dataDir)
	}
	if strings.TrimSpace(ipcPath) != "" {
		cfg.Integrations.OpenDesign.IPCPath = strings.TrimSpace(ipcPath)
	}
	cfg.Integrations.OpenDesign.Mode = config.ConfigModeMCP
	cfg.Integrations.OpenDesign.Fallback = "design-doc-only"
}

func findOpenDesignDaemonCLI(probe platform.SystemProbe) (cliPath string, dataDir string) {
	for _, root := range openDesignRootCandidates(probe) {
		if root == "" {
			continue
		}
		candidate := filepath.Join(root, "apps", "daemon", "dist", "cli.js")
		if info, err := probe.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, filepath.Join(root, ".od")
		}
	}
	return "", ""
}

func openDesignRootCandidates(probe platform.SystemProbe) []string {
	info := platform.DetectPlatform(probe)
	home := info.HomeDir
	values := []string{
		strings.TrimSpace(probe.Getenv("OPENDESIGN_ROOT")),
		strings.TrimSpace(probe.Getenv("OPEN_DESIGN_ROOT")),
		strings.TrimSpace(probe.Getenv("OD_ROOT")),
	}
	if home != "" {
		values = append(values,
			filepath.Join(home, "Documents", "TOOLS", "APPs", "open-design"),
			filepath.Join(home, "Documents", "TOOLS", "open-design"),
			filepath.Join(home, "Documents", "open-design"),
			filepath.Join(home, "Developer", "open-design"),
			filepath.Join(home, "dev", "open-design"),
			filepath.Join(home, "open-design"),
		)
	}
	return values
}

func findOpenDesignNode(probe platform.SystemProbe) string {
	for _, envKey := range []string{"OPENDESIGN_NODE_PATH", "NODE_BINARY", "NODE_PATH"} {
		candidate := strings.TrimSpace(probe.Getenv(envKey))
		if candidate == "" {
			continue
		}
		if info, err := probe.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	for _, binary := range []string{"node", "node.exe"} {
		if path, err := probe.LookPath(binary); err == nil && strings.TrimSpace(path) != "" {
			return path
		}
	}
	for _, candidate := range []string{"/opt/homebrew/bin/node", "/usr/local/bin/node", "/usr/bin/node"} {
		if info, err := probe.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}

func defaultOpenDesignIPCPath(probe platform.SystemProbe) string {
	for _, key := range []string{"OPENDESIGN_SIDECAR_IPC_PATH", "OD_SIDECAR_IPC_PATH"} {
		if value := strings.TrimSpace(probe.Getenv(key)); value != "" {
			return value
		}
	}
	if probe != nil && probe.GOOS() == "windows" {
		return ""
	}
	return "/tmp/open-design/ipc/default/daemon.sock"
}

func isOpenDesignCLICommand(command string) bool {
	base := strings.ToLower(strings.TrimSuffix(filepath.Base(strings.TrimSpace(command)), ".exe"))
	return base == "od" || base == "open-design"
}
