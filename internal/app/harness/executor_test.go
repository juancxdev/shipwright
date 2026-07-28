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
	if !ArtifactExists(CommunicationPolicyFile) {
		t.Fatalf("expected generated file %s", CommunicationPolicyFile)
	}
	assertFileContains(t, "AGENTS.md", "Shipwright Project Instructions")
	assertFileContains(t, "AGENTS.md", ".harness/communication-policy.md")
	assertFileContains(t, CommunicationPolicyFile, "neutral professional Spanish")
	assertFileContains(t, CommunicationPolicyFile, "overrides global/personal assistant personality settings")
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
		CommunicationPolicyFile,
		filepath.Join(".opencode", "AGENTS.md"),
		filepath.Join(".harness", "bin", "shipwright"),
		filepath.Join(".harness", "bin", "shipwright.cmd"),
		filepath.Join(".opencode", "opencode.json"),
		filepath.Join(".opencode", "mcp", "package.json"),
		filepath.Join(".opencode", "mcp", "stitch-proxy.mjs"),
		filepath.Join(".opencode", "mcp", "README.md"),
		filepath.Join(".opencode", "mcp", ".gitignore"),
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
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "OpenCode integration")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), ".harness/communication-policy.md")
	assertFileContains(t, CommunicationPolicyFile, "neutral professional Spanish")
	assertFileContains(t, CommunicationPolicyFile, "Do not use regional dialects")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), ".harness/project-profile.md")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), ".harness/tdd-policy.md")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), ".harness/skill-digests.md")
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
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "{file:./AGENTS.md}")
	assertFileNotContains(t, filepath.Join(".opencode", "opencode.json"), "\"instructions\"")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "{file:./agents/product-owner.md}")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "\"stitch_*\"")
	assertFileContains(t, filepath.Join(".opencode", "mcp", "package.json"), "@google/stitch-sdk")
	assertFileContains(t, filepath.Join(".opencode", "mcp", "stitch-proxy.mjs"), "StitchProxy")
	assertFileContains(t, filepath.Join(".opencode", "mcp", "stitch-proxy.mjs"), ".harness")
	assertFileContains(t, filepath.Join(".opencode", "mcp", "README.md"), "npm install --prefix .opencode/mcp")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "Shipwright Orchestrator Autopilot")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "Senior SDLC Delivery Orchestrator")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "The orchestrator is a coordinator, not an executor")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "Always delegate active role work")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "If OpenCode task/subagent delegation is unavailable, stop")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "delegation unavailable")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "execution, not delegation")
	assertFileNotContains(t, filepath.Join(".opencode", "AGENTS.md"), "otherwise follow that role's")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "shipwright start")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "Do not ask the user to run `next`")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "shipwright approve scope")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "treat `installed_no_active_canvas` as **unverified**")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "Treat role ownership as a routing contract")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "`ui-ux-designer`: UI/UX design")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "Stitch/OpenDesign/OpenPencil/Figma-like provider work")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "`frontend-engineer`: frontend implementation in HTML, CSS, JavaScript, TypeScript, React")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "`backend-engineer`: backend implementation, APIs, services")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "`qa-security-reviewer`: QA verification, regression review, security review")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "Provider/canvas work belongs ONLY to `ui-ux-designer`")
	assertFileContains(t, filepath.Join(".opencode", "AGENTS.md"), "Never delegate OpenDesign/Stitch/OpenPencil tasks to `frontend-engineer`")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "\"open-pencil_*\"")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "\"open-design_list_projects\"")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "\"open-design_create_artifact\"")
	assertFileContains(t, filepath.Join(".opencode", "opencode.json"), "\"opendesign_*\"")
	assertFileContains(t, filepath.Join(".harness", "bin", "shipwright"), "../shipwright")
	assertFileContains(t, filepath.Join(".harness", "bin", "shipwright"), "SHIPWRIGHT_BIN")
	assertFileContains(t, filepath.Join(".harness", "bin", "shipwright"), "shell aliases are not visible")
	if ArtifactExists(filepath.Join(".harness", "bin", "loom")) || ArtifactExists(filepath.Join(".harness", "bin", "loom.cmd")) {
		t.Fatal("legacy loom wrappers should not be generated for unreleased Shipwright")
	}
	wrapper, err := os.ReadFile(filepath.Join(".harness", "bin", "shipwright"))
	if err != nil {
		t.Fatalf("read shipwright wrapper: %v", err)
	}
	if strings.Contains(string(wrapper), "LOOM_BIN") {
		t.Fatal("shipwright wrapper should not reference LOOM_BIN")
	}
	if info, err := os.Stat(filepath.Join(".harness", "bin", "shipwright")); err != nil {
		t.Fatalf("stat shipwright wrapper: %v", err)
	} else if info.Mode()&0111 == 0 {
		t.Fatalf("shipwright wrapper should be executable, mode=%s", info.Mode())
	}
	if ArtifactExists("opencode.json") {
		t.Fatal("root opencode.json should not be generated; OpenCode config belongs in .opencode/opencode.json")
	}
	if ArtifactExists("AGENTS.md") {
		t.Fatal("root AGENTS.md should not be generated for OpenCode; instructions belong in .opencode/AGENTS.md")
	}
	assertFileContains(t, filepath.Join(".opencode", "agents", "product-owner.md"), "mode: subagent")
	assertFileContains(t, filepath.Join(".opencode", "agents", "product-owner.md"), "senior professional identity")
	assertFileContains(t, filepath.Join(".opencode", "agents", "product-owner.md"), "communication-policy.md")
	assertFileContains(t, filepath.Join(".opencode", "agents", "product-owner.md"), "project-profile.md")
	assertFileContains(t, filepath.Join(".opencode", "agents", "frontend-engineer.md"), "tdd-policy.md")
	assertFileContains(t, filepath.Join(".opencode", "agents", "frontend-engineer.md"), "provider work must be handled by `ui-ux-designer`")
	assertFileContains(t, filepath.Join(".opencode", "agents", "ui-ux-designer.md"), "Stitch")
	assertFileContains(t, filepath.Join(".opencode", "agents", "ui-ux-designer.md"), "Do not delegate Stitch, OpenDesign, OpenPencil")
	assertFileContains(t, filepath.Join(".opencode", "skills", "ui-ux-designer", "SKILL.md"), "installed_no_active_canvas")
	assertFileContains(t, filepath.Join(".opencode", "skills", "ui-ux-designer", "SKILL.md"), "Responsive & Accessibility QA")
	assertFileContains(t, filepath.Join(".opencode", "skills", "ui-ux-designer", "SKILL.md"), "Existing Web Baseline Fidelity Gate")
	assertFileContains(t, filepath.Join(".opencode", "skills", "existing-web-to-openpencil", "SKILL.md"), "fidelity-report.md")
	assertFileContains(t, filepath.Join(".opencode", "skills", "existing-web-to-openpencil", "SKILL.md"), "Section matrix per route/viewport")
	assertFileContains(t, filepath.Join(".opencode", "skills", "stitch-generate-design", "SKILL.md"), "Google Stitch")
	assertFileContains(t, filepath.Join(".opencode", "skills", "stitch-generate-design", "SKILL.md"), "STITCH_API_KEY")
	assertFileContains(t, filepath.Join(".opencode", "skills", "opendesign-generate-artifact", "SKILL.md"), "ARTIFACT_MANIFEST_REQUIRED")
	assertFileContains(t, filepath.Join(".opencode", "skills", "canvas-generate-design", "SKILL.md"), "Build like a designer")
	assertFileContains(t, filepath.Join(".opencode", "skills", "canvas-generate-design", "SKILL.md"), "Figma-inspired canvas discipline")
	assertFileContains(t, filepath.Join(".opencode", "skills", "openpencil-generate-design", "SKILL.md"), "open-pencil_get_current_page")
	assertFileContains(t, filepath.Join(".opencode", "skills", "openpencil-generate-design", "SKILL.md"), "save-status.md")
	assertFileContains(t, filepath.Join(".opencode", "skills", "openpencil-generate-design", "SKILL.md"), "Inspect the current page/canvas before drawing")
	assertFileContains(t, filepath.Join(".opencode", "skills", "design-code-component-map", "SKILL.md"), "code-component-map.md")
	assertFileContains(t, filepath.Join(".opencode", "skills", "design-code-component-map", "SKILL.md"), "Official Figma Code Connect mode")
	assertFileContains(t, filepath.Join(".opencode", "skills", "ui-ux-designer", "SKILL.md"), "No component extends outside its frame/canvas")
	assertFileContains(t, filepath.Join(".opencode", "skills", "product-owner", "SKILL.md"), "name: product-owner")
	assertFileContains(t, filepath.Join(".opencode", "skills", "product-owner", "SKILL.md"), "Senior Product Owner and Business Analyst")
	assertFileContains(t, filepath.Join(".opencode", "skills", "frontend-engineer", "SKILL.md"), "Senior Frontend Engineer")
	assertFileContains(t, filepath.Join(".opencode", "skills", "_shared", "agent-common.md"), "Professional Identity Contract")
	assertFileContains(t, filepath.Join(".opencode", "skills", "frontend-design", "SKILL.md"), "name: frontend-design")
	assertFileContains(t, filepath.Join(".opencode", "skills", "accessibility", "SKILL.md"), "Keyboard first")
	assertFileContains(t, filepath.Join(".opencode", "skills", "responsive-layout-systems", "SKILL.md"), "390×844")

	status, err := OpenCodeExecutorAdapter{}.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.Configured || len(status.Missing) != 0 {
		t.Fatalf("status = %+v", status)
	}
}

func TestOpenCodeProviderToolsAreDesignerOnly(t *testing.T) {
	designerTools := opencodeToolsForAgent("ui-ux-designer")
	for _, key := range []string{"open-design_*", "opendesign_*", "open_design_*", "stitch_*", "open-pencil_*"} {
		if designerTools[key] != true {
			t.Fatalf("ui-ux-designer tool %s = %+v, want true", key, designerTools[key])
		}
	}

	frontendTools := opencodeToolsForAgent("frontend-engineer")
	for _, key := range []string{"open-design_*", "opendesign_*", "open_design_*", "stitch_*", "open-pencil_*"} {
		if _, ok := frontendTools[key]; ok {
			t.Fatalf("frontend-engineer must not receive provider tool %s", key)
		}
	}

	frontendPermissionMap := opencodePermissionMapForAgent("frontend-engineer", opencodePermission{Edit: "allow", Bash: "ask"})
	for _, key := range []string{"open-design_*", "opendesign_*", "open_design_*"} {
		if _, ok := frontendPermissionMap[key]; ok {
			t.Fatalf("frontend-engineer permission map must not allow provider tool %s", key)
		}
	}
}

func TestExecutorDoesNotOverwriteExistingCommunicationPolicy(t *testing.T) {
	withTempWorkingDir(t)

	customPolicy := "# Communication Policy\n\nCustom project tone.\n"
	if err := WriteFile(CommunicationPolicyFile, customPolicy); err != nil {
		t.Fatalf("write custom policy: %v", err)
	}

	if _, err := GenerateExecutor(ExecutorOpenCode); err != nil {
		t.Fatalf("GenerateExecutor: %v", err)
	}

	data, err := os.ReadFile(CommunicationPolicyFile)
	if err != nil {
		t.Fatalf("read communication policy: %v", err)
	}
	if string(data) != customPolicy {
		t.Fatalf("communication policy overwritten:\n%s", string(data))
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

func TestOpenCodeExecutorUsesFastForBalancedWhenDefaultImplicit(t *testing.T) {
	withTempWorkingDir(t)

	cfg := DefaultPortableConfig()
	ApplyOpenCodeModelOverrides(cfg, OpenCodeModelOverrides{
		FastModel:      "opencode-go/deepseek-v4-flash",
		ReasoningModel: "opencode-go/deepseek-v4-flash",
	})
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}
	if err := SaveModelPolicy(DefaultModelPolicy(cfg.Executors.OpenCode)); err != nil {
		t.Fatalf("save model policy: %v", err)
	}

	if _, err := GenerateExecutor(ExecutorOpenCode); err != nil {
		t.Fatalf("GenerateExecutor: %v", err)
	}

	configPath := filepath.Join(".opencode", "opencode.json")
	for _, agent := range []string{"product-owner", "frontend-engineer", "backend-engineer", "technical-lead", "qa-security-reviewer", "shipwright-orchestrator"} {
		assertFileContains(t, configPath, `"`+agent+`"`)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read %s: %v", configPath, err)
	}
	if strings.Contains(string(data), "anthropic/claude-sonnet-4-20250514") {
		t.Fatal("OpenCode config should not contain implicit Anthropic default when fast/reasoning models were explicitly set")
	}
	assertFileContains(t, configPath, "opencode-go/deepseek-v4-flash")
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

func TestOpenCodeExecutorIncludesStitchMCPWhenLocalSecretConfigured(t *testing.T) {
	withTempWorkingDir(t)

	if err := SaveLocalSecret("STITCH_API_KEY", "test-key"); err != nil {
		t.Fatalf("SaveLocalSecret: %v", err)
	}
	cfg := DefaultPortableConfig()

	json := opencodeJSONWithConfig(DefaultOpenCodeExecutorConfig(), cfg)
	if !strings.Contains(json, "\"mcp\"") || !strings.Contains(json, "\"stitch\"") || !strings.Contains(json, ".opencode/mcp/stitch-proxy.mjs") {
		t.Fatalf("opencode json missing stitch mcp config:\n%s", json)
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

func assertFileNotContains(t *testing.T, path string, needle string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if strings.Contains(string(data), needle) {
		t.Fatalf("%s contains %q", path, needle)
	}
}

func TestOpenCodeExecutorIncludesOpenDesignMCPWhenConfigured(t *testing.T) {
	withTempWorkingDir(t)

	command := filepath.Join(t.TempDir(), "node")
	if err := os.WriteFile(command, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatalf("write node: %v", err)
	}
	cfg := DefaultPortableConfig()
	cfg.Integrations.OpenDesign.MCPCommand = command
	cfg.Integrations.OpenDesign.MCPArgs = []string{"/tools/open-design/apps/daemon/dist/cli.js", "mcp"}
	cfg.Integrations.OpenDesign.DaemonURL = "http://127.0.0.1:7377"
	cfg.Integrations.OpenDesign.DataDir = "/tools/open-design/.od"
	cfg.Integrations.OpenDesign.IPCPath = "/tmp/open-design/ipc/default/daemon.sock"
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	if _, err := GenerateExecutor(ExecutorOpenCode); err != nil {
		t.Fatalf("GenerateExecutor: %v", err)
	}

	configPath := filepath.Join(".opencode", "opencode.json")
	assertFileContains(t, configPath, "\"open-design\"")
	assertFileContains(t, configPath, ".opencode/mcp/open-design.sh")
	assertFileContains(t, configPath, "\"open-design_*\"")
	wrapperPath := filepath.Join(".opencode", "mcp", "open-design.sh")
	assertFileContains(t, wrapperPath, "OD_DATA_DIR")
	assertFileContains(t, wrapperPath, "OD_DAEMON_URL")
	assertFileContains(t, wrapperPath, "http://127.0.0.1:7377")
	assertFileContains(t, wrapperPath, "OD_SIDECAR_IPC_PATH")
	assertFileContains(t, wrapperPath, "exec '")
}

func TestOpenCodeExecutorRespectsDisabledOpenPencilWhenOpenDesignSelected(t *testing.T) {
	withTempWorkingDir(t)

	node := filepath.Join(t.TempDir(), "node")
	if err := os.WriteFile(node, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatalf("write node: %v", err)
	}
	openpencil := filepath.Join(t.TempDir(), "openpencil-mcp")
	if err := os.WriteFile(openpencil, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatalf("write openpencil: %v", err)
	}

	cfg := DefaultPortableConfig()
	cfg.Integrations.OpenDesign.MCPCommand = node
	cfg.Integrations.OpenDesign.MCPArgs = []string{"/tools/open-design/apps/daemon/dist/cli.js", "mcp"}
	cfg.Integrations.OpenPencil.MCPCommand = openpencil
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	integrations := DefaultIntegrations()
	integrations.Stitch.Enabled = false
	integrations.Stitch.Status = "disabled_via_cli"
	integrations.OpenDesign.Enabled = true
	integrations.OpenDesign.Status = "available"
	integrations.OpenPencil.Enabled = false
	integrations.OpenPencil.Status = "installed_no_active_canvas"
	if err := integrations.Save(); err != nil {
		t.Fatalf("save integrations: %v", err)
	}

	if _, err := GenerateExecutor(ExecutorOpenCode); err != nil {
		t.Fatalf("GenerateExecutor: %v", err)
	}

	configPath := filepath.Join(".opencode", "opencode.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read %s: %v", configPath, err)
	}
	json := string(data)
	if !strings.Contains(json, "\"open-design\"") || !strings.Contains(json, "\"open-design_*\"") {
		t.Fatalf("expected OpenDesign MCP/tools in OpenCode config:\n%s", json)
	}
	if strings.Contains(json, "\"open-pencil\"") || strings.Contains(json, "\"open-pencil_*\"") {
		t.Fatalf("OpenCode config must not expose OpenPencil when openpencil.enabled=false:\n%s", json)
	}

	assertFileContains(t, filepath.Join(".opencode", "agents", "ui-ux-designer.md"), "Provider selection is authoritative")
	assertFileContains(t, filepath.Join(".opencode", "agents", "ui-ux-designer.md"), "If `openpencil.enabled=false`, never call `open-pencil_*`")
}
