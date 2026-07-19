package harness

import "testing"

func TestValidatePortableConfigReportsSemanticErrors(t *testing.T) {
	cfg := DefaultPortableConfig()
	cfg.Version = "999"
	cfg.Health.TimeoutMillis = -1
	cfg.Integrations.Engram.Mode = "magic"
	cfg.Integrations.Engram.HealthURL = "ftp://localhost/health"
	cfg.Integrations.OpenPencil.MCPServerPath = "relative/mcp-server.cjs"

	issues := ValidatePortableConfig(cfg)
	errorsCount, warningsCount := CountConfigIssueSeverities(issues)
	if errorsCount < 4 {
		t.Fatalf("errors = %d, issues=%+v", errorsCount, issues)
	}
	if warningsCount < 1 {
		t.Fatalf("warnings = %d, issues=%+v", warningsCount, issues)
	}
}

func TestValidatePortableConfigAcceptsDefaultConfig(t *testing.T) {
	cfg := DefaultPortableConfig()
	issues := ValidatePortableConfig(cfg)
	if len(issues) != 0 {
		t.Fatalf("default config should validate, got %+v", issues)
	}
}

func TestValidatePortableConfigAcceptsWindowsAbsolutePaths(t *testing.T) {
	cfg := DefaultPortableConfig()
	cfg.PlatformOverrides["windows"] = PortableIntegrationsConfig{
		Engram: PortableEngramConfig{BinaryPath: `C:\\Tools\\engram.exe`},
		OpenPencil: PortableOpenPencilConfig{
			MCPServerPath: `C:\\Tools\\OpenPencil\\mcp-server.cjs`,
		},
	}
	issues := ValidatePortableConfig(cfg)
	if len(issues) != 0 {
		t.Fatalf("windows absolute paths should validate, got %+v", issues)
	}
}

func TestValidatePortableConfigWarnsUnknownOpenCodeAgentModel(t *testing.T) {
	cfg := DefaultPortableConfig()
	cfg.Executors.OpenCode.AgentModels["typo-agent"] = "opencode-go/deepseek-v4-flash"

	issues := ValidatePortableConfig(cfg)
	_, warningsCount := CountConfigIssueSeverities(issues)
	if warningsCount == 0 {
		t.Fatalf("expected warning for unknown OpenCode agent, got %+v", issues)
	}
}
