package harness

import (
	"os"
	"path/filepath"
	"testing"
)

func withTempWorkingDir(t *testing.T) string {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
	return dir
}

func TestLoadEffectivePortableConfigDefaultsWhenMissing(t *testing.T) {
	withTempWorkingDir(t)

	cfg, err := LoadEffectivePortableConfig(fakeProbe{goos: "linux"})
	if err != nil {
		t.Fatalf("LoadEffectivePortableConfig: %v", err)
	}

	if cfg.Version != PortableConfigVersion {
		t.Fatalf("version = %s", cfg.Version)
	}
	if cfg.Integrations.Engram.Fallback != DecisionsFile {
		t.Fatalf("engram fallback = %s", cfg.Integrations.Engram.Fallback)
	}
	if cfg.Integrations.OpenPencil.Fallback != "design-doc-only" {
		t.Fatalf("openpencil fallback = %s", cfg.Integrations.OpenPencil.Fallback)
	}
}

func TestLoadEffectivePortableConfigAppliesPlatformOverride(t *testing.T) {
	withTempWorkingDir(t)

	cfg := DefaultPortableConfig()
	cfg.PlatformOverrides["windows"] = PortableIntegrationsConfig{
		Engram: PortableEngramConfig{BinaryPath: `C:\\Tools\\engram.exe`},
		OpenPencil: PortableOpenPencilConfig{
			MCPServerPath: `C:\\Tools\\OpenPencil\\mcp-server.cjs`,
		},
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	effective, err := LoadEffectivePortableConfig(fakeProbe{goos: "windows"})
	if err != nil {
		t.Fatalf("LoadEffectivePortableConfig: %v", err)
	}

	if effective.Integrations.Engram.BinaryPath != `C:\\Tools\\engram.exe` {
		t.Fatalf("engram binary = %s", effective.Integrations.Engram.BinaryPath)
	}
	if effective.Integrations.OpenPencil.MCPServerPath != `C:\\Tools\\OpenPencil\\mcp-server.cjs` {
		t.Fatalf("openpencil mcp = %s", effective.Integrations.OpenPencil.MCPServerPath)
	}
}

func TestLoadEffectivePortableConfigEnvWinsOverFileAndPlatform(t *testing.T) {
	withTempWorkingDir(t)

	cfg := DefaultPortableConfig()
	cfg.Integrations.Engram.BinaryPath = "/file/engram"
	cfg.PlatformOverrides["linux"] = PortableIntegrationsConfig{
		Engram: PortableEngramConfig{BinaryPath: "/platform/engram"},
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	effective, err := LoadEffectivePortableConfig(fakeProbe{
		goos: "linux",
		env: map[string]string{
			"ENGRAM_BINARY": "/env/engram",
		},
	})
	if err != nil {
		t.Fatalf("LoadEffectivePortableConfig: %v", err)
	}

	if effective.Integrations.Engram.BinaryPath != "/env/engram" {
		t.Fatalf("engram binary = %s", effective.Integrations.Engram.BinaryPath)
	}
}

func TestDetectionUsesPortableConfigWithoutEnv(t *testing.T) {
	withTempWorkingDir(t)

	engramPath := "/opt/shipwright/bin/engram"
	mcpPath := "/opt/openpencil/mcp-server.cjs"
	cfg := DefaultPortableConfig()
	cfg.Integrations.Engram.BinaryPath = engramPath
	cfg.Integrations.OpenPencil.MCPServerPath = mcpPath

	probe := fakeProbe{
		goos: "linux",
		statMap: map[string]fakeFileInfo{
			engramPath: {name: filepath.Base(engramPath)},
			mcpPath:    {name: filepath.Base(mcpPath)},
		},
	}

	engram := DetectEngramWithConfig(probe, cfg)
	if !engram.Available || engram.Path != engramPath || engram.PathKind != DetectionPathBinary {
		t.Fatalf("engram detection = %+v", engram)
	}

	openpencil := DetectOpenPencilWithConfig(probe, cfg)
	if !openpencil.Installed || openpencil.Path != mcpPath || openpencil.PathKind != DetectionPathMCPServer {
		t.Fatalf("openpencil detection = %+v", openpencil)
	}
}

func TestIntegrationsApplyPortableConfig(t *testing.T) {
	cfg := DefaultPortableConfig()
	cfg.Integrations.Engram.BinaryPath = "/usr/bin/engram"
	cfg.Integrations.Engram.HealthURL = "http://127.0.0.1:9999/health"
	cfg.Integrations.OpenPencil.AppPath = "/Applications/OpenPencil.app"
	cfg.Integrations.OpenPencil.MCPServerPath = "/Applications/OpenPencil.app/Contents/Resources/mcp-server.cjs"

	integrations := DefaultIntegrations()
	integrations.ApplyPortableConfig(cfg)

	if integrations.Engram.BinaryPath != cfg.Integrations.Engram.BinaryPath {
		t.Fatalf("engram binary = %s", integrations.Engram.BinaryPath)
	}
	if integrations.Engram.HealthURL != cfg.Integrations.Engram.HealthURL {
		t.Fatalf("engram health = %s", integrations.Engram.HealthURL)
	}
	if integrations.OpenPencil.AppPath != cfg.Integrations.OpenPencil.AppPath {
		t.Fatalf("openpencil app = %s", integrations.OpenPencil.AppPath)
	}
	if integrations.OpenPencil.MCPServerPath != cfg.Integrations.OpenPencil.MCPServerPath {
		t.Fatalf("openpencil mcp = %s", integrations.OpenPencil.MCPServerPath)
	}
}

func TestLoadEffectivePortableConfigEnvOverridesHealthTimeout(t *testing.T) {
	withTempWorkingDir(t)

	cfg, err := LoadEffectivePortableConfig(fakeProbe{
		goos: "linux",
		env:  map[string]string{"SHIPWRIGHT_HEALTH_TIMEOUT_MS": "250"},
	})
	if err != nil {
		t.Fatalf("LoadEffectivePortableConfig: %v", err)
	}

	if cfg.Health.TimeoutMillis != 250 {
		t.Fatalf("timeout = %d, want 250", cfg.Health.TimeoutMillis)
	}
}

func TestLoadEffectivePortableConfigEnvOverridesOpenCodeModels(t *testing.T) {
	withTempWorkingDir(t)

	cfg, err := LoadEffectivePortableConfig(fakeProbe{
		goos: "linux",
		env: map[string]string{
			"SHIPWRIGHT_OPENCODE_DEFAULT_MODEL":   "opencode-go/deepseek-v4-flash",
			"SHIPWRIGHT_OPENCODE_REASONING_MODEL": "openai/gpt-5.5",
			"SHIPWRIGHT_OPENCODE_FAST_MODEL":      "opencode-go/deepseek-v4-flash",
			"SHIPWRIGHT_OPENCODE_AGENT_MODELS":    "technical-lead=openai/gpt-5.5,product-owner=opencode-go/deepseek-v4-flash",
		},
	})
	if err != nil {
		t.Fatalf("LoadEffectivePortableConfig: %v", err)
	}

	if cfg.Executors.OpenCode.ReasoningModel != "openai/gpt-5.5" {
		t.Fatalf("reasoning model = %s", cfg.Executors.OpenCode.ReasoningModel)
	}
	if cfg.Executors.OpenCode.AgentModels["technical-lead"] != "openai/gpt-5.5" {
		t.Fatalf("technical-lead model = %s", cfg.Executors.OpenCode.AgentModels["technical-lead"])
	}
}
