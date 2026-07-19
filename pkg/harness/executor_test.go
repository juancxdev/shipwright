package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenericExecutorGeneratesAgentsMD(t *testing.T) {
	withTempWorkingDir(t)

	result, err := GenerateExecutor(ExecutorGeneric)
	if err != nil {
		t.Fatalf("GenerateExecutor: %v", err)
	}
	if result.Name != ExecutorGeneric {
		t.Fatalf("name = %s", result.Name)
	}
	assertFileContains(t, "AGENTS.md", "Shipwright Project Instructions")
	assertFileContains(t, ExecutorStateFile, ExecutorGeneric)

	status, err := GenericExecutorAdapter{}.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.Configured || len(status.Missing) != 0 {
		t.Fatalf("status = %+v", status)
	}
}

func TestOpenCodeExecutorGeneratesSupportedFiles(t *testing.T) {
	withTempWorkingDir(t)

	result, err := GenerateExecutor(ExecutorOpenCode)
	if err != nil {
		t.Fatalf("GenerateExecutor: %v", err)
	}
	if result.Name != ExecutorOpenCode {
		t.Fatalf("name = %s", result.Name)
	}

	expected := []string{
		"AGENTS.md",
		filepath.Join(".harness", "bin", "shipwright"),
		filepath.Join(".harness", "bin", "shipwright.cmd"),
		filepath.Join(".opencode", "opencode.json"),
		filepath.Join(".opencode", "agents", "product-owner.md"),
		filepath.Join(".opencode", "commands", "shipwright-status.md"),
		filepath.Join(".opencode", "skills", "product-owner", "SKILL.md"),
		filepath.Join(".opencode", "skills", "_shared", "agent-common.md"),
		ExecutorStateFile,
	}
	for _, file := range expected {
		if !ArtifactExists(file) {
			t.Fatalf("expected generated file %s", file)
		}
	}
	assertFileContains(t, "AGENTS.md", "OpenCode integration")
	assertFileContains(t, "AGENTS.md", ".harness/project-profile.md")
	assertFileContains(t, "AGENTS.md", ".harness/tdd-policy.md")
	assertFileContains(t, "AGENTS.md", ".harness/skill-digests.md")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "https://opencode.ai/config.json")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "\"default_agent\"")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "\"shipwright-orchestrator\"")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "\"agent\"")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "\"command\"")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "\"shipwright-orchestrator\"")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "\"model\"")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "anthropic/claude-sonnet-4-20250514")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "\"product-owner\"")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "\"shipwright-status\"")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "{file:../AGENTS.md}")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "{file:./agents/product-owner.md}")
	assertFileContains(t, "AGENTS.md", "Shipwright Orchestrator Autopilot")
	assertFileContains(t, "AGENTS.md", ".harness/bin/shipwright start")
	assertFileContains(t, "AGENTS.md", "Do not ask the user to run `next`")
	assertFileContains(t, "AGENTS.md", ".harness/bin/shipwright approve scope")
	assertFileContains(t, "AGENTS.md", "treat `installed_no_active_canvas` as **unverified**")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "\"open-pencil_*\"")
	assertFileContains(t, filepath.Join(".harness", "bin", "shipwright"), "../shipwright")
	assertFileContains(t, filepath.Join(".harness", "bin", "shipwright"), "SHIPWRIGHT_BIN")
	if info, err := os.Stat(filepath.Join(".harness", "bin", "shipwright")); err != nil {
		t.Fatalf("stat shipwright wrapper: %v", err)
	} else if info.Mode()&0111 == 0 {
		t.Fatalf("shipwright wrapper should be executable, mode=%s", info.Mode())
	}
	if ArtifactExists("opencode.json") {
		t.Fatal("root opencode.json should not be generated; OpenCode config belongs in .opencode/opencode.json")
	}
	assertFileContains(t, filepath.Join(".opencode", "agents", "product-owner.md"), "mode: subagent")
	assertFileContains(t, filepath.Join(".opencode", "agents", "product-owner.md"), "project-profile.md")
	assertFileContains(t, filepath.Join(".opencode", "agents", "frontend-engineer.md"), "tdd-policy.md")
	assertFileContains(t, filepath.Join(".opencode", "agents", "ui-ux-designer.md"), "OpenPencil MCP validation")
	assertFileContains(t, filepath.Join(".opencode", "skills", "ui-ux-designer", "SKILL.md"), "installed_no_active_canvas")
	assertFileContains(t, filepath.Join(".opencode", "skills", "ui-ux-designer", "SKILL.md"), "Responsive & Accessibility QA")
	assertFileContains(t, filepath.Join(".opencode", "skills", "ui-ux-designer", "SKILL.md"), "No component extends outside its frame/canvas")
	assertFileContains(t, filepath.Join(".opencode", "skills", "product-owner", "SKILL.md"), "name: product-owner")

	status, err := OpenCodeExecutorAdapter{}.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.Configured || len(status.Missing) != 0 {
		t.Fatalf("status = %+v", status)
	}
}

func TestOpenCodeExecutorUsesConfiguredModels(t *testing.T) {
	withTempWorkingDir(t)

	cfg := DefaultPortableConfig()
	cfg.Executors.OpenCode.DefaultModel = "opencode-go/deepseek-v4-flash"
	cfg.Executors.OpenCode.ReasoningModel = "openai/gpt-5.5"
	cfg.Executors.OpenCode.FastModel = "opencode-go/deepseek-v4-flash"
	cfg.Executors.OpenCode.AgentModels["product-owner"] = "custom/product-owner-model"
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	if _, err := GenerateExecutor(ExecutorOpenCode); err != nil {
		t.Fatalf("GenerateExecutor: %v", err)
	}

	configPath := filepath.Join(".opencode", "opencode.json")
	assertFileContains(t, configPath, "openai/gpt-5.5")
	assertFileContains(t, configPath, "opencode-go/deepseek-v4-flash")
	assertFileContains(t, configPath, "custom/product-owner-model")
}

func TestOpenCodeExecutorIncludesOpenPencilMCPWhenConfigured(t *testing.T) {
	withTempWorkingDir(t)

	mcpServer := filepath.Join(t.TempDir(), "mcp-server.cjs")
	if err := os.WriteFile(mcpServer, []byte("console.log('mcp')\n"), 0644); err != nil {
		t.Fatalf("write mcp server: %v", err)
	}
	cfg := DefaultPortableConfig()
	cfg.Integrations.OpenPencil.MCPServerPath = mcpServer

	json := opencodeJSONWithConfig(DefaultOpenCodeExecutorConfig(), cfg)
	if !strings.Contains(json, "\"mcp\"") || !strings.Contains(json, "\"open-pencil\"") || !strings.Contains(json, mcpServer) {
		t.Fatalf("opencode json missing open-pencil mcp config:\n%s", json)
	}
}

func TestUnknownExecutorFails(t *testing.T) {
	if _, err := GetExecutorAdapter("unknown"); err == nil {
		t.Fatal("expected unknown executor error")
	}
}

func assertFileContains(t *testing.T, path string, needle string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(data), needle) {
		t.Fatalf("%s does not contain %q", path, needle)
	}
}
