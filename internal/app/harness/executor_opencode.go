package harness

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type OpenCodeExecutorAdapter struct{}

func (OpenCodeExecutorAdapter) Name() string { return ExecutorOpenCode }

func (OpenCodeExecutorAdapter) Description() string {
	return "OpenCode project bootstrap: .opencode/AGENTS.md, .opencode/opencode.json, agents, commands, and skills."
}

func (OpenCodeExecutorAdapter) Generate() (*ExecutorGenerateResult, error) {
	result := &ExecutorGenerateResult{Name: ExecutorOpenCode}
	if err := ensureTrackedFile(CommunicationPolicyFile, DefaultCommunicationPolicyMarkdown(), result); err != nil {
		return nil, err
	}
	if err := writeTrackedFile(opencodeAgentsPath(), opencodeAgentsMD(), result); err != nil {
		return nil, err
	}
	if err := writeExecutableTrackedFile(filepath.Join(".harness", "bin", "shipwright"), opencodeShipwrightWrapperSH(), result); err != nil {
		return nil, err
	}
	if err := writeTrackedFile(filepath.Join(".harness", "bin", "shipwright.cmd"), opencodeShipwrightWrapperCMD(), result); err != nil {
		return nil, err
	}
	if err := writeTrackedFile(openCodeConfigPath(), opencodeJSON(), result); err != nil {
		return nil, err
	}
	if err := writeTrackedFile(filepath.Join(".opencode", "mcp", "package.json"), opencodeStitchMCPPackageJSON(), result); err != nil {
		return nil, err
	}
	if err := writeTrackedFile(filepath.Join(".opencode", "mcp", "stitch-proxy.mjs"), opencodeStitchMCPProxy(), result); err != nil {
		return nil, err
	}
	if err := writeExecutableTrackedFile(filepath.Join(".opencode", "mcp", "open-design.sh"), opencodeOpenDesignMCPWrapper(), result); err != nil {
		return nil, err
	}
	if err := writeTrackedFile(filepath.Join(".opencode", "mcp", "README.md"), opencodeStitchMCPReadme(), result); err != nil {
		return nil, err
	}
	if err := writeTrackedFile(filepath.Join(".opencode", "mcp", ".gitignore"), "node_modules/\n", result); err != nil {
		return nil, err
	}
	if err := writeTrackedFile(filepath.Join(".opencode", "skills", "_shared", "agent-common.md"), AgentCommonProtocol, result); err != nil {
		return nil, err
	}
	for _, skill := range AllAgentSkills() {
		if err := writeTrackedFile(opencodeAgentPath(skill.Name), opencodeAgentMarkdown(skill), result); err != nil {
			return nil, err
		}
		if err := writeTrackedFile(opencodeSkillPath(skill.Name), skill.Content, result); err != nil {
			return nil, err
		}
	}
	for _, skill := range AllCuratedSkills() {
		if err := writeTrackedFile(opencodeSkillPath(skill.Name), skill.Content, result); err != nil {
			return nil, err
		}
	}
	for _, command := range opencodeCommands() {
		if err := writeTrackedFile(filepath.Join(".opencode", "commands", command.Filename), command.Content, result); err != nil {
			return nil, err
		}
	}
	result.Message = "OpenCode executor generated. OpenCode will read .opencode/AGENTS.md and .opencode/opencode.json with .opencode/agents, .opencode/commands, and .opencode/skills."
	return result, nil
}

func (OpenCodeExecutorAdapter) Status() (*ExecutorStatus, error) {
	files := []string{
		CommunicationPolicyFile,
		opencodeAgentsPath(),
		filepath.Join(".harness", "bin", "shipwright"),
		filepath.Join(".harness", "bin", "shipwright.cmd"),
		openCodeConfigPath(),
		filepath.Join(".opencode", "mcp", "package.json"),
		filepath.Join(".opencode", "mcp", "stitch-proxy.mjs"),
		filepath.Join(".opencode", "mcp", "open-design.sh"),
		filepath.Join(".opencode", "mcp", "README.md"),
		filepath.Join(".opencode", "mcp", ".gitignore"),
		filepath.Join(".opencode", "skills", "_shared", "agent-common.md"),
	}
	for _, skill := range AllAgentSkills() {
		files = append(files, opencodeAgentPath(skill.Name), opencodeSkillPath(skill.Name))
	}
	for _, skill := range AllCuratedSkills() {
		files = append(files, opencodeSkillPath(skill.Name))
	}
	for _, command := range opencodeCommands() {
		files = append(files, filepath.Join(".opencode", "commands", command.Filename))
	}
	status := requiredStatus(ExecutorOpenCode, files)
	if ArtifactExists("AGENTS.md") {
		status.Warnings = append(status.Warnings, "root AGENTS.md exists; Shipwright writes OpenCode instructions to .opencode/AGENTS.md to keep executor assets together.")
	}
	if ArtifactExists("opencode.json") {
		status.Warnings = append(status.Warnings, "root opencode.json exists; Shipwright writes OpenCode project config to .opencode/opencode.json to keep executor assets together.")
	}
	if !ArtifactExists(".harness/state.json") {
		status.Warnings = append(status.Warnings, "Shipwright harness is not initialized; run shipwright init before using OpenCode executor files.")
	}
	return status, nil
}

func opencodeJSON() string {
	modelConfig := DefaultOpenCodeExecutorConfig()
	var cfg *PortableConfig
	if loaded, err := LoadEffectivePortableConfig(RealSystemProbe{}); err == nil {
		cfg = loaded
		modelConfig = loaded.Executors.OpenCode
	}
	return opencodeJSONWithConfig(modelConfig, cfg)
}

func opencodeJSONWithModels(modelConfig PortableOpenCodeExecutorConfig) string {
	return opencodeJSONWithConfig(modelConfig, nil)
}

func opencodeJSONWithConfig(modelConfig PortableOpenCodeExecutorConfig, cfg *PortableConfig) string {
	modelConfig.Normalize()
	integrations := loadOptionalIntegrations()
	agents := map[string]any{
		"shipwright-orchestrator": map[string]any{
			"mode":        "primary",
			"description": "Shipwright lifecycle orchestrator - reads harness state and delegates work to Shipwright role agents",
			"model":       ResolveOpenCodeModelWithPolicy("shipwright-orchestrator", modelConfig, LoadOrDefaultModelPolicy(modelConfig)),
			"prompt":      "{file:./AGENTS.md}",
			"permission": map[string]any{
				"edit":     "ask",
				"bash":     "ask",
				"question": "allow",
				"task":     opencodeTaskPermissions(),
			},
			"tools": map[string]any{
				"read":     true,
				"write":    false,
				"edit":     false,
				"bash":     true,
				"question": true,
				"task":     true,
			},
		},
	}
	for _, skill := range AllAgentSkills() {
		permission := opencodePermissionForAgent(skill.Name)
		agents[skill.Name] = map[string]any{
			"mode":        "subagent",
			"description": extractSkillDescription(skill.Content),
			"model":       ResolveOpenCodeModelWithPolicy(skill.Name, modelConfig, LoadOrDefaultModelPolicy(modelConfig)),
			"prompt":      fmt.Sprintf("{file:./agents/%s.md}", skill.Name),
			"permission":  opencodePermissionMapForAgentWithIntegrations(skill.Name, permission, integrations),
			"tools":       opencodeToolsForAgentWithIntegrations(skill.Name, integrations),
		}
	}

	payload := map[string]any{
		"$schema":       "https://opencode.ai/config.json",
		"default_agent": "shipwright-orchestrator",
		"agent":         agents,
		"command":       opencodeCommandConfig(),
	}
	if mcp := opencodeMCPConfig(cfg, integrations); len(mcp) > 0 {
		payload["mcp"] = mcp
	}
	data, _ := json.MarshalIndent(payload, "", "  ")
	return string(data) + "\n"
}

func loadOptionalIntegrations() *Integrations {
	integrations, err := LoadIntegrations()
	if err != nil {
		return nil
	}
	return integrations
}

func opencodeMCPConfig(cfg *PortableConfig, integrations *Integrations) map[string]any {
	if cfg == nil {
		return nil
	}
	mcp := map[string]any{}

	stitch := DetectStitchWithConfig(RealSystemProbe{}, cfg)
	if providerEnabled(integrations, DesignModeStitch) && stitch.Available {
		mcp["stitch"] = map[string]any{
			"type":    "local",
			"command": opencodeStitchMCPCommand(),
			"enabled": true,
		}
	}

	opendesign := DetectOpenDesignWithConfig(RealSystemProbe{}, cfg)
	if providerEnabled(integrations, DesignModeOpenDesign) && opendesign.Configured && opendesign.Installed {
		mcp["open-design"] = map[string]any{
			"type":    "local",
			"command": opencodeOpenDesignMCPCommand(),
			"enabled": true,
		}
	}

	detected := DetectOpenPencilWithConfig(RealSystemProbe{}, cfg)
	if providerEnabled(integrations, DesignModeOpenPencil) && detected.Installed && detected.Path != "" {
		var command []string
		switch detected.PathKind {
		case DetectionPathBinary:
			command = []string{detected.Path}
		case DetectionPathMCPServer:
			command = []string{"node", detected.Path, "--stdio"}
		}
		if len(command) > 0 {
			mcp["open-pencil"] = map[string]any{
				"type":    "local",
				"command": command,
				"enabled": true,
			}
		}
	}

	return mcp
}

func providerEnabled(integrations *Integrations, provider string) bool {
	if integrations == nil {
		return true
	}
	switch provider {
	case DesignModeStitch:
		return integrations.Stitch.Enabled
	case DesignModeOpenDesign:
		return integrations.OpenDesign.Enabled
	case DesignModeOpenPencil:
		return integrations.OpenPencil.Enabled
	default:
		return false
	}
}

func opencodeStitchMCPPackageJSON() string {
	return `{
  "private": true,
  "type": "module",
  "dependencies": {
    "@google/stitch-sdk": "^0.3.5"
  }
}
`
}

func opencodeStitchMCPProxy() string {
	return `#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { StitchProxy } from "@google/stitch-sdk";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";

function loadLocalEnv(filePath) {
  if (!fs.existsSync(filePath)) return;
  const content = fs.readFileSync(filePath, "utf8");
  for (const rawLine of content.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith("#") || !line.includes("=")) continue;
    const [rawKey, ...rest] = line.split("=");
    const key = rawKey.trim();
    let value = rest.join("=").trim();
    value = value.replace(/^['"]|['"]$/g, "");
    if (key && value && !process.env[key]) {
      process.env[key] = value;
    }
  }
}

const cwd = process.cwd();
loadLocalEnv(path.join(cwd, ".harness", "secrets.local.env"));
loadLocalEnv(path.join(cwd, "..", ".harness", "secrets.local.env"));

if (process.env.SHIPWRIGHT_STITCH_MCP_URL && !process.env.STITCH_HOST) {
  process.env.STITCH_HOST = process.env.SHIPWRIGHT_STITCH_MCP_URL;
}

const apiKey = process.env.STITCH_API_KEY;
if (!apiKey) {
  console.error("Shipwright Stitch MCP: STITCH_API_KEY not found. Set it in the environment or .harness/secrets.local.env.");
  process.exit(1);
}

const proxy = new StitchProxy({ apiKey });
const transport = new StdioServerTransport();
await proxy.start(transport);
`
}

func opencodeStitchMCPReadme() string {
	return "# Shipwright MCP Adapters\n\n" +
		"Shipwright generates local MCP helpers so OpenCode can expose design tools to the UI/UX Designer.\n\n" +
		"## Stitch\n\n" +
		"Install once per project:\n\n" +
		"```bash\n" +
		"npm install --prefix .opencode/mcp\n" +
		"```\n\n" +
		"Then restart OpenCode and verify:\n\n" +
		"```bash\n" +
		"opencode mcp list\n" +
		"```\n\n" +
		"You should see a connected server named `stitch`. In the agent chat, ask:\n\n" +
		"```txt\n" +
		"List my Stitch projects.\n" +
		"```\n\n" +
		"Credentials:\n\n" +
		"- Preferred local-project secret: `.harness/secrets.local.env`\n" +
		"- Environment fallback: `STITCH_API_KEY`\n\n" +
		"Do not commit `.harness/secrets.local.env`.\n\n" +
		"## OpenDesign\n\n" +
		"Configure once through Shipwright, not by editing OpenCode JSON manually:\n\n" +
		"```bash\n" +
		"shipwright integrations configure opendesign --command /path/to/node --arg /path/to/open-design/apps/daemon/dist/cli.js --arg mcp --daemon-url http://127.0.0.1:7377 --data-dir /path/to/open-design/.od --ipc-path /tmp/open-design/ipc/default/daemon.sock\n" +
		"shipwright executor generate opencode\n" +
		"```\n\n" +
		"`--ipc-path` is a Shipwright config flag, not an OpenDesign daemon flag. Do not run `cli.js daemon --ipc-path`.\n\n" +
		"Then restart OpenCode and verify `open-design` with `opencode mcp list`.\n"
}

func opencodeStitchMCPInstallHint() string {
	return `npm install --prefix .opencode/mcp`
}

func opencodeStitchMCPRequiredPackage() string {
	return "@google/stitch-sdk"
}

func opencodeStitchMCPServerName() string {
	return "stitch"
}

func opencodeStitchMCPToolPattern() string {
	return "stitch_*"
}

func opencodeStitchMCPCommand() []string {
	return []string{"node", ".opencode/mcp/stitch-proxy.mjs"}
}

func opencodeOpenDesignMCPCommand() []string {
	return []string{".opencode/mcp/open-design.sh"}
}

func opencodeOpenDesignMCPWrapper() string {
	cfg, err := LoadEffectivePortableConfig(RealSystemProbe{})
	if err != nil || cfg == nil || strings.TrimSpace(cfg.Integrations.OpenDesign.MCPCommand) == "" {
		return `#!/usr/bin/env sh
set -eu
echo "Shipwright OpenDesign MCP is not configured. Run: shipwright integrations configure opendesign --help" >&2
exit 127
`
	}
	command := shellQuote(cfg.Integrations.OpenDesign.MCPCommand)
	args := make([]string, 0, len(cfg.Integrations.OpenDesign.MCPArgs))
	for _, arg := range cfg.Integrations.OpenDesign.MCPArgs {
		if strings.TrimSpace(arg) != "" {
			args = append(args, shellQuote(arg))
		}
	}
	var builder strings.Builder
	builder.WriteString("#!/usr/bin/env sh\nset -eu\n")
	if strings.TrimSpace(cfg.Integrations.OpenDesign.DataDir) != "" {
		builder.WriteString("export OD_DATA_DIR=")
		builder.WriteString(shellQuote(cfg.Integrations.OpenDesign.DataDir))
		builder.WriteString("\n")
	}
	if strings.TrimSpace(cfg.Integrations.OpenDesign.DaemonURL) != "" {
		builder.WriteString("export OD_DAEMON_URL=")
		builder.WriteString(shellQuote(strings.TrimRight(strings.TrimSpace(cfg.Integrations.OpenDesign.DaemonURL), "/")))
		builder.WriteString("\n")
	}
	if strings.TrimSpace(cfg.Integrations.OpenDesign.IPCPath) != "" {
		builder.WriteString("export OD_SIDECAR_IPC_PATH=")
		builder.WriteString(shellQuote(cfg.Integrations.OpenDesign.IPCPath))
		builder.WriteString("\n")
	}
	builder.WriteString("exec ")
	builder.WriteString(command)
	for _, arg := range args {
		builder.WriteByte(' ')
		builder.WriteString(arg)
	}
	builder.WriteString("\n")
	return builder.String()
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func opencodeShipwrightWrapperSH() string {
	return `#!/usr/bin/env sh
set -eu
if [ -n "${SHIPWRIGHT_BIN:-}" ] && [ -x "${SHIPWRIGHT_BIN}" ]; then
  exec "${SHIPWRIGHT_BIN}" "$@"
fi
if [ -x "../shipwright" ]; then
  exec ../shipwright "$@"
fi
if [ -x "../shipwright.exe" ]; then
  exec ../shipwright.exe "$@"
fi
if command -v shipwright >/dev/null 2>&1; then
  exec shipwright "$@"
fi
if command -v harness >/dev/null 2>&1; then
  exec harness "$@"
fi
echo "Shipwright CLI not found. Install shipwright globally, set SHIPWRIGHT_BIN, or keep the binary one directory above this project. Note: shell aliases are not visible to this wrapper." >&2
exit 127
`
}

func opencodeShipwrightWrapperCMD() string {
	return `@echo off
if not "%SHIPWRIGHT_BIN%"=="" if exist "%SHIPWRIGHT_BIN%" "%SHIPWRIGHT_BIN%" %* & exit /b %errorlevel%
if exist ..\shipwright.exe ..\shipwright.exe %* & exit /b %errorlevel%
if exist ..\shipwright ..\shipwright %* & exit /b %errorlevel%
where shipwright >nul 2>nul
if %errorlevel%==0 shipwright %* & exit /b %errorlevel%
where harness >nul 2>nul
if %errorlevel%==0 harness %* & exit /b %errorlevel%
echo Shipwright CLI not found. Install shipwright globally, set SHIPWRIGHT_BIN, or keep shipwright.exe one directory above this project. 1>&2
exit /b 127
`
}

func opencodeTaskPermissions() map[string]any {
	permissions := map[string]any{"*": "deny"}
	for _, skill := range AllAgentSkills() {
		permissions[skill.Name] = "allow"
	}
	return permissions
}

func opencodeToolsForAgent(name string) map[string]any {
	return opencodeToolsForAgentWithIntegrations(name, nil)
}

func opencodeToolsForAgentWithIntegrations(name string, integrations *Integrations) map[string]any {
	switch name {
	case "ui-ux-designer":
		tools := map[string]any{
			"read":  true,
			"write": true,
			"edit":  true,
			"bash":  false,
			"task":  false,
		}
		if providerEnabled(integrations, DesignModeStitch) {
			tools["stitch_*"] = true
		}
		if providerEnabled(integrations, DesignModeOpenDesign) {
			tools["open-design_*"] = true
			tools["opendesign_*"] = true
			tools["open_design_*"] = true
			for _, tool := range opencodeOpenDesignToolNames() {
				tools[tool] = true
			}
		}
		if providerEnabled(integrations, DesignModeOpenPencil) {
			tools["open-pencil_*"] = true
		}
		return tools
	case "product-owner", "project-manager":
		return map[string]any{"read": true, "write": true, "edit": true, "bash": false, "task": false}
	case "qa-security-reviewer":
		return map[string]any{"read": true, "write": true, "edit": true, "bash": true, "task": false}
	case "technical-lead", "frontend-engineer", "backend-engineer":
		return map[string]any{"read": true, "write": true, "edit": true, "bash": true, "task": false}
	default:
		return map[string]any{"read": true, "write": true, "edit": false, "bash": false, "task": false}
	}
}

func opencodeOpenDesignToolNames() []string {
	base := []string{
		"list_projects",
		"get_active_context",
		"get_artifact",
		"get_project",
		"get_file",
		"search_files",
		"list_files",
		"create_artifact",
	}
	prefixes := []string{"open-design", "opendesign", "open_design"}
	tools := make([]string, 0, len(base)*(len(prefixes)+1))
	tools = append(tools, base...)
	for _, prefix := range prefixes {
		for _, name := range base {
			tools = append(tools, prefix+"_"+name)
		}
	}
	return tools
}

func opencodePermissionMapForAgent(name string, permission opencodePermission) map[string]any {
	return opencodePermissionMapForAgentWithIntegrations(name, permission, nil)
}

func opencodePermissionMapForAgentWithIntegrations(name string, permission opencodePermission, integrations *Integrations) map[string]any {
	out := map[string]any{
		"edit": permission.Edit,
		"bash": permission.Bash,
	}
	if name == "ui-ux-designer" && providerEnabled(integrations, DesignModeOpenDesign) {
		out["open-design_*"] = "allow"
		out["opendesign_*"] = "allow"
		out["open_design_*"] = "allow"
		for _, tool := range opencodeOpenDesignToolNames() {
			out[tool] = "allow"
		}
	}
	return out
}

func opencodeCommandConfig() map[string]any {
	commands := map[string]any{}
	for _, command := range opencodeCommands() {
		name := strings.TrimSuffix(command.Filename, ".md")
		commands[name] = map[string]any{
			"description": extractCommandDescription(command.Content),
			"template":    extractCommandTemplate(command.Content),
			"agent":       "shipwright-orchestrator",
			"subtask":     false,
		}
	}
	return commands
}

func extractCommandDescription(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "description:") {
			return strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "description:")), "\"")
		}
	}
	return "Shipwright command"
}

func extractCommandTemplate(content string) string {
	parts := strings.SplitN(content, "---", 3)
	if len(parts) == 3 {
		return strings.TrimSpace(parts[2])
	}
	return strings.TrimSpace(content)
}

const opencodeAgentsAppendixTemplate = "templates/project/harness/executor/opencode-agents-appendix.md"
const opencodeAgentTemplate = "templates/project/harness/executor/opencode-agent.md"
const opencodeAgentUIDesignerExtraTemplate = "templates/project/harness/executor/opencode-agent-ui-ux-designer-extra.md"
const opencodeAgentFrontendExtraTemplate = "templates/project/harness/executor/opencode-agent-frontend-engineer-extra.md"

func opencodeAgentsMD() string {
	appendix, err := RenderTemplate(opencodeAgentsAppendixTemplate, nil)
	if err != nil {
		panic(err)
	}
	return genericAgentsMD() + "\n" + appendix
}

func opencodeAgentMarkdown(skill AgentSkill) string {
	desc := extractSkillDescription(skill.Content)
	permission := opencodePermissionForAgent(skill.Name)
	content, err := RenderTemplate(opencodeAgentTemplate, RenderVars{
		"description":     yamlQuote(desc),
		"permission_edit": permission.Edit,
		"permission_bash": permission.Bash,
		"agent_name":      skill.Name,
		"skill_name":      skill.Name,
		"role_extra":      opencodeAgentExtraMarkdown(skill.Name),
	})
	if err != nil {
		panic(err)
	}
	return content
}

func opencodeAgentExtraMarkdown(name string) string {
	var template string
	switch name {
	case "ui-ux-designer":
		template = opencodeAgentUIDesignerExtraTemplate
	case "frontend-engineer":
		template = opencodeAgentFrontendExtraTemplate
	default:
		return ""
	}
	content, err := RenderTemplate(template, nil)
	if err != nil {
		panic(err)
	}
	return content
}

type opencodePermission struct{ Edit, Bash string }

func opencodePermissionForAgent(name string) opencodePermission {
	switch name {
	case "product-owner", "project-manager", "ui-ux-designer", "qa-security-reviewer", "technical-lead", "frontend-engineer", "backend-engineer":
		return opencodePermission{Edit: "allow", Bash: "ask"}
	default:
		return opencodePermission{Edit: "ask", Bash: "ask"}
	}
}

type opencodeCommand struct{ Filename, Content string }

func opencodeCommands() []opencodeCommand {
	return []opencodeCommand{
		{Filename: "shipwright-status.md", Content: opencodeCommandMarkdown("Shipwright status", "Run `shipwright status` (fallback: `.harness/bin/shipwright status`), summarize current phase, active gates, missing artifacts, and the safest next Shipwright command. Do not modify files.", "plan")},
		{Filename: "shipwright-active-agent.md", Content: opencodeCommandMarkdown("Shipwright active agent", "Run `shipwright agents active` (fallback: `.harness/bin/shipwright agents active`), identify the active Shipwright agent, then read `shipwright agents run <agent>` before proposing work. Do not modify files unless the active role allows it.", "plan")},
		{Filename: "shipwright-next.md", Content: opencodeCommandMarkdown("Shipwright next gate", "Run `shipwright status` and determine whether `shipwright next` is safe. If this is a non-approval internal transition and gates/evidence are satisfied, run `shipwright next` yourself. If an approval gate is blocking, present the artifact and ask the user to approve or request changes. Fallback: `.harness/bin/shipwright` only if global `shipwright` is unavailable.", "plan")},
		{Filename: "shipwright-doctor.md", Content: opencodeCommandMarkdown("Shipwright doctor", "Run `shipwright doctor` (fallback: `.harness/bin/shipwright doctor`) and summarize blocking errors, warnings, fallbacks, and concrete fixes. Do not edit config unless the user explicitly asks for `shipwright doctor --fix`.", "plan")},
	}
}

func opencodeCommandMarkdown(description, prompt, agent string) string {
	return fmt.Sprintf("---\ndescription: %s\nagent: %s\n---\n\n%s\n", yamlQuote(description), agent, prompt)
}

func extractSkillDescription(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "description:") {
			desc := strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			return strings.Trim(desc, "\"")
		}
	}
	return "Shipwright role agent"
}

func yamlQuote(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return "\"" + value + "\""
}
