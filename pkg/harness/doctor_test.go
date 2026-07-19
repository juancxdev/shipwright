package harness

import (
	"testing"
	"time"
)

func TestRunDoctorMissingPortableConfigUsesFallbacks(t *testing.T) {
	withTempWorkingDir(t)

	report, err := RunDoctor(fakeProbe{goos: "linux", goarch: "amd64"})
	if err != nil {
		t.Fatalf("RunDoctor: %v", err)
	}

	if report.ConfigExists {
		t.Fatal("config should be reported as missing")
	}
	if report.ConfigLoaded != true {
		t.Fatal("defaults should be loaded when config is missing")
	}
	if report.Summary.Errors != 0 {
		t.Fatalf("errors = %d, want 0", report.Summary.Errors)
	}
	if report.Summary.Warnings == 0 {
		t.Fatal("expected warning for missing portable config")
	}
	if report.Engram.Fallback != DecisionsFile {
		t.Fatalf("engram fallback = %s", report.Engram.Fallback)
	}
	if report.OpenPencil.Fallback != "design-doc-only" {
		t.Fatalf("openpencil fallback = %s", report.OpenPencil.Fallback)
	}
}

func TestRunDoctorWarnsOnBrokenConfiguredPaths(t *testing.T) {
	withTempWorkingDir(t)

	cfg := DefaultPortableConfig()
	cfg.Integrations.Engram.BinaryPath = "/missing/engram"
	cfg.Integrations.OpenPencil.MCPServerPath = "/missing/openpencil/mcp-server.cjs"
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	report, err := RunDoctor(fakeProbe{
		goos: "linux",
		statMap: map[string]fakeFileInfo{
			PortableConfigFile: {name: "config.json"},
		},
	})
	if err != nil {
		t.Fatalf("RunDoctor: %v", err)
	}

	if !report.ConfigExists || !report.ConfigLoaded {
		t.Fatalf("config exists=%v loaded=%v", report.ConfigExists, report.ConfigLoaded)
	}
	if report.Summary.Errors != 0 {
		t.Fatalf("errors = %d", report.Summary.Errors)
	}
	if report.Summary.Warnings < 2 {
		t.Fatalf("warnings = %d, want at least 2", report.Summary.Warnings)
	}
	if report.Engram.Status != DetectionNotInstalled || !report.Engram.Configured {
		t.Fatalf("engram = %+v", report.Engram)
	}
	if report.OpenPencil.Status != DetectionNotInstalled || !report.OpenPencil.Configured {
		t.Fatalf("openpencil = %+v", report.OpenPencil)
	}
}

func TestRunDoctorHealthyWhenIntegrationsAvailable(t *testing.T) {
	withTempWorkingDir(t)

	engramPath := "/usr/local/bin/engram"
	mcpPath := "/Applications/OpenPencil.app/Contents/Resources/mcp-server.cjs"
	cfg := DefaultPortableConfig()
	cfg.Integrations.Engram.BinaryPath = engramPath
	cfg.Integrations.OpenPencil.MCPServerPath = mcpPath
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	report, err := RunDoctor(fakeProbe{
		goos: "darwin",
		env:  map[string]string{"OPENPENCIL_CANVAS_ACTIVE": "true"},
		statMap: map[string]fakeFileInfo{
			PortableConfigFile: {name: "config.json"},
			engramPath:         {name: "engram"},
			mcpPath:            {name: "mcp-server.cjs"},
		},
	})
	if err != nil {
		t.Fatalf("RunDoctor: %v", err)
	}

	if !report.Summary.Healthy {
		t.Fatalf("doctor should be healthy: %+v", report.Summary)
	}
	if !report.Engram.Available {
		t.Fatalf("engram = %+v", report.Engram)
	}
	if !report.OpenPencil.Available || !report.OpenPencil.Active {
		t.Fatalf("openpencil = %+v", report.OpenPencil)
	}
}

func TestRunDoctorReportsBrokenConfigAsError(t *testing.T) {
	withTempWorkingDir(t)
	if err := WriteFile(PortableConfigFile, "{ broken json"); err != nil {
		t.Fatalf("write broken config: %v", err)
	}

	report, err := RunDoctor(fakeProbe{
		goos: "linux",
		statMap: map[string]fakeFileInfo{
			PortableConfigFile: {name: "config.json"},
		},
	})
	if err != nil {
		t.Fatalf("RunDoctor should return report even with config error: %v", err)
	}

	if report.Summary.Errors != 1 {
		t.Fatalf("errors = %d, want 1; checks=%+v", report.Summary.Errors, report.Checks)
	}
	if report.Summary.Healthy {
		t.Fatal("broken config should not be healthy")
	}
}

type fakeHealthProbe struct {
	httpResults    map[string]HealthResult
	commandResults map[string]HealthResult
}

func (f fakeHealthProbe) CheckHTTP(url string, timeout time.Duration) HealthResult {
	if result, ok := f.httpResults[url]; ok {
		return result
	}
	return HealthResult{Name: "http", Checked: true, Status: HealthStatusUnhealthy, Endpoint: url, Detail: "unexpected url"}
}

func (f fakeHealthProbe) CheckCommand(name string, args []string, timeout time.Duration) HealthResult {
	key := name
	if len(args) > 0 {
		key = name + " " + args[len(args)-1]
	}
	if result, ok := f.commandResults[key]; ok {
		return result
	}
	return HealthResult{Name: name, Checked: true, Status: HealthStatusUnhealthy, Detail: "unexpected command"}
}

func TestRunDoctorReportsHealthyRealChecks(t *testing.T) {
	withTempWorkingDir(t)

	engramPath := "/usr/local/bin/engram"
	mcpPath := "/Applications/OpenPencil.app/Contents/Resources/mcp-server.cjs"
	cfg := DefaultPortableConfig()
	cfg.Integrations.Engram.BinaryPath = engramPath
	cfg.Integrations.OpenPencil.MCPServerPath = mcpPath
	cfg.Integrations.Engram.HealthURL = "http://localhost:9999/health"
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	report, err := RunDoctorWithHealth(fakeProbe{
		goos: "darwin",
		env:  map[string]string{"OPENPENCIL_CANVAS_ACTIVE": "true"},
		statMap: map[string]fakeFileInfo{
			PortableConfigFile: {name: "config.json"},
			engramPath:         {name: "engram"},
			mcpPath:            {name: "mcp-server.cjs"},
		},
	}, fakeHealthProbe{
		httpResults: map[string]HealthResult{
			"http://localhost:9999/health": {Checked: true, Healthy: true, Status: HealthStatusHealthy, Endpoint: "http://localhost:9999/health", Detail: "200 OK", LatencyMS: 3},
		},
		commandResults: map[string]HealthResult{
			"node " + mcpPath: {Checked: true, Healthy: true, Status: HealthStatusHealthy, Detail: "mcp-server-load-ok", LatencyMS: 5},
		},
	})
	if err != nil {
		t.Fatalf("RunDoctorWithHealth: %v", err)
	}

	if !report.EngramHealth.Healthy {
		t.Fatalf("engram health = %+v", report.EngramHealth)
	}
	if !report.OpenPencilHealth.Healthy {
		t.Fatalf("openpencil health = %+v", report.OpenPencilHealth)
	}
	if report.Summary.Errors != 0 || report.Summary.Warnings != 0 {
		t.Fatalf("summary = %+v", report.Summary)
	}
}

func TestRunDoctorWarnsWhenHealthChecksFail(t *testing.T) {
	withTempWorkingDir(t)

	engramPath := "/usr/local/bin/engram"
	mcpPath := "/Applications/OpenPencil.app/Contents/Resources/mcp-server.cjs"
	cfg := DefaultPortableConfig()
	cfg.Integrations.Engram.BinaryPath = engramPath
	cfg.Integrations.OpenPencil.MCPServerPath = mcpPath
	cfg.Integrations.Engram.HealthURL = "http://localhost:9999/health"
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	report, err := RunDoctorWithHealth(fakeProbe{
		goos: "darwin",
		env:  map[string]string{"OPENPENCIL_CANVAS_ACTIVE": "true"},
		statMap: map[string]fakeFileInfo{
			PortableConfigFile: {name: "config.json"},
			engramPath:         {name: "engram"},
			mcpPath:            {name: "mcp-server.cjs"},
		},
	}, fakeHealthProbe{
		httpResults: map[string]HealthResult{
			"http://localhost:9999/health": {Checked: true, Status: HealthStatusUnhealthy, Endpoint: "http://localhost:9999/health", Detail: "connection refused"},
		},
		commandResults: map[string]HealthResult{
			"node " + mcpPath: {Checked: true, Status: HealthStatusUnhealthy, Detail: "module load failed"},
		},
	})
	if err != nil {
		t.Fatalf("RunDoctorWithHealth: %v", err)
	}

	if report.Summary.Errors != 0 {
		t.Fatalf("health failure should not be blocking: %+v", report.Summary)
	}
	if report.Summary.Warnings < 2 {
		t.Fatalf("warnings = %d, want at least 2", report.Summary.Warnings)
	}
}

func TestRunDoctorReportsConfigValidationErrors(t *testing.T) {
	withTempWorkingDir(t)

	cfg := DefaultPortableConfig()
	cfg.Version = "999"
	cfg.Integrations.Engram.HealthURL = "not-a-url"
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}
	// Save() normalizes but must not erase explicitly invalid non-empty fields.

	report, err := RunDoctorWithHealth(fakeProbe{
		goos: "linux",
		statMap: map[string]fakeFileInfo{
			PortableConfigFile: {name: "config.json"},
		},
	}, fakeHealthProbe{})
	if err != nil {
		t.Fatalf("RunDoctorWithHealth: %v", err)
	}

	if len(report.ConfigIssues) == 0 {
		t.Fatal("expected config issues")
	}
	if report.Summary.Errors == 0 {
		t.Fatalf("expected blocking config errors, summary=%+v checks=%+v", report.Summary, report.Checks)
	}
}

func TestApplyDoctorFixesCreatesMissingConfig(t *testing.T) {
	withTempWorkingDir(t)

	result, err := ApplyDoctorFixes(RealSystemProbe{})
	if err != nil {
		t.Fatalf("ApplyDoctorFixes: %v", err)
	}
	if !result.Applied {
		t.Fatalf("expected fix applied: %+v", result)
	}
	if !ArtifactExists(PortableConfigFile) {
		t.Fatal("expected portable config to be created")
	}
}

func TestApplyDoctorFixesBacksUpCorruptConfig(t *testing.T) {
	withTempWorkingDir(t)
	if err := WriteFile(PortableConfigFile, "{ broken json"); err != nil {
		t.Fatalf("write corrupt config: %v", err)
	}

	result, err := ApplyDoctorFixes(RealSystemProbe{})
	if err != nil {
		t.Fatalf("ApplyDoctorFixes: %v", err)
	}
	if !result.Applied || result.Backup == "" {
		t.Fatalf("expected backup fix, got %+v", result)
	}
	if !ArtifactExists(result.Backup) {
		t.Fatalf("expected backup file %s", result.Backup)
	}
	if _, err := LoadPortableConfig(); err != nil {
		t.Fatalf("replacement config should load: %v", err)
	}
}

func TestApplyDoctorFixesNormalizesConfig(t *testing.T) {
	withTempWorkingDir(t)
	if err := WriteFile(PortableConfigFile, `{"version":"1","artifact_root":".","integrations":{"engram":{},"openpencil":{}}}`); err != nil {
		t.Fatalf("write partial config: %v", err)
	}

	result, err := ApplyDoctorFixes(RealSystemProbe{})
	if err != nil {
		t.Fatalf("ApplyDoctorFixes: %v", err)
	}
	if !result.Applied {
		t.Fatalf("expected normalization: %+v", result)
	}
	cfg, err := LoadPortableConfig()
	if err != nil {
		t.Fatalf("LoadPortableConfig: %v", err)
	}
	if cfg.Health.TimeoutMillis != DefaultHealthTimeoutMillis {
		t.Fatalf("timeout = %d", cfg.Health.TimeoutMillis)
	}
	if cfg.Integrations.Engram.Fallback != DecisionsFile {
		t.Fatalf("fallback = %s", cfg.Integrations.Engram.Fallback)
	}
}

func TestApplyDoctorFixesRepairsSafeSemanticIssues(t *testing.T) {
	withTempWorkingDir(t)
	cfg := DefaultPortableConfig()
	cfg.Version = "999"
	cfg.Health.TimeoutMillis = -10
	cfg.Integrations.Engram.Mode = "magic"
	cfg.Integrations.Engram.HealthURL = "ftp://localhost/health"
	cfg.Integrations.OpenPencil.Mode = "other"
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	result, err := ApplyDoctorFixes(RealSystemProbe{})
	if err != nil {
		t.Fatalf("ApplyDoctorFixes: %v", err)
	}
	if !result.Applied {
		t.Fatalf("expected fixes: %+v", result)
	}

	repaired, err := LoadPortableConfigRaw()
	if err != nil {
		t.Fatalf("LoadPortableConfigRaw: %v", err)
	}
	if repaired.Version != PortableConfigVersion {
		t.Fatalf("version = %s", repaired.Version)
	}
	if repaired.Integrations.Engram.Mode != ConfigModeMCP {
		t.Fatalf("engram mode = %s", repaired.Integrations.Engram.Mode)
	}
	if repaired.Integrations.OpenPencil.Mode != ConfigModeMCP {
		t.Fatalf("openpencil mode = %s", repaired.Integrations.OpenPencil.Mode)
	}
	if repaired.Integrations.Engram.HealthURL != "http://localhost:7437/health" {
		t.Fatalf("health url = %s", repaired.Integrations.Engram.HealthURL)
	}
	if issues := ValidatePortableConfig(repaired); len(issues) != 0 {
		t.Fatalf("repaired config should validate, got %+v", issues)
	}
}
