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
	return "OpenCode project bootstrap: AGENTS.md, .opencode/opencode.json, agents, commands, and skills."
}

func (OpenCodeExecutorAdapter) Generate() (*ExecutorGenerateResult, error) {
	result := &ExecutorGenerateResult{Name: ExecutorOpenCode}
	if err := writeTrackedFile("AGENTS.md", opencodeAgentsMD(), result); err != nil {
		return nil, err
	}
	if err := writeExecutableTrackedFile(filepath.Join(".harness", "bin", "shipwright"), opencodeShipwrightWrapperSH(), result); err != nil {
		return nil, err
	}
	if err := writeTrackedFile(filepath.Join(".harness", "bin", "shipwright.cmd"), opencodeShipwrightWrapperCMD(), result); err != nil {
		return nil, err
	}
	// Legacy wrappers keep older generated OpenCode prompts working during the LOOM -> Shipwright rename.
	if err := writeExecutableTrackedFile(filepath.Join(".harness", "bin", "loom"), opencodeShipwrightWrapperSH(), result); err != nil {
		return nil, err
	}
	if err := writeTrackedFile(filepath.Join(".harness", "bin", "loom.cmd"), opencodeShipwrightWrapperCMD(), result); err != nil {
		return nil, err
	}
	if err := writeTrackedFile(openCodeConfigPath(), opencodeJSON(), result); err != nil {
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
	for _, command := range opencodeCommands() {
		if err := writeTrackedFile(filepath.Join(".opencode", "commands", command.Filename), command.Content, result); err != nil {
			return nil, err
		}
	}
	result.Message = "OpenCode executor generated. OpenCode will read AGENTS.md and .opencode/opencode.json with .opencode/agents, .opencode/commands, and .opencode/skills."
	return result, nil
}

func (OpenCodeExecutorAdapter) Status() (*ExecutorStatus, error) {
	files := []string{"AGENTS.md", filepath.Join(".harness", "bin", "shipwright"), filepath.Join(".harness", "bin", "shipwright.cmd"), filepath.Join(".harness", "bin", "loom"), filepath.Join(".harness", "bin", "loom.cmd"), openCodeConfigPath(), filepath.Join(".opencode", "skills", "_shared", "agent-common.md")}
	for _, skill := range AllAgentSkills() {
		files = append(files, opencodeAgentPath(skill.Name), opencodeSkillPath(skill.Name))
	}
	for _, command := range opencodeCommands() {
		files = append(files, filepath.Join(".opencode", "commands", command.Filename))
	}
	status := requiredStatus(ExecutorOpenCode, files)
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
	agents := map[string]any{
		"shipwright-orchestrator": map[string]any{
			"mode":        "primary",
			"description": "Shipwright lifecycle orchestrator - reads harness state and delegates work to Shipwright role agents",
			"model":       ResolveOpenCodeModel("shipwright-orchestrator", modelConfig),
			"prompt":      "{file:../AGENTS.md}",
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
			"model":       ResolveOpenCodeModel(skill.Name, modelConfig),
			"prompt":      fmt.Sprintf("{file:./agents/%s.md}", skill.Name),
			"permission": map[string]any{
				"edit": permission.Edit,
				"bash": permission.Bash,
			},
			"tools": opencodeToolsForAgent(skill.Name),
		}
	}

	payload := map[string]any{
		"$schema":       "https://opencode.ai/config.json",
		"default_agent": "shipwright-orchestrator",
		"instructions":  []string{"../AGENTS.md"},
		"agent":         agents,
		"command":       opencodeCommandConfig(),
	}
	if mcp := opencodeMCPConfig(cfg); len(mcp) > 0 {
		payload["mcp"] = mcp
	}
	data, _ := json.MarshalIndent(payload, "", "  ")
	return string(data) + "\n"
}

func opencodeMCPConfig(cfg *PortableConfig) map[string]any {
	if cfg == nil {
		return nil
	}
	detected := DetectOpenPencilWithConfig(RealSystemProbe{}, cfg)
	if !detected.Installed || detected.Path == "" {
		return nil
	}
	var command []string
	switch detected.PathKind {
	case DetectionPathBinary:
		command = []string{detected.Path}
	case DetectionPathMCPServer:
		command = []string{"node", detected.Path, "--stdio"}
	default:
		return nil
	}
	return map[string]any{
		"open-pencil": map[string]any{
			"type":    "local",
			"command": command,
			"enabled": true,
		},
	}
}

func opencodeShipwrightWrapperSH() string {
	return `#!/usr/bin/env sh
set -eu
if [ -n "${SHIPWRIGHT_BIN:-}" ] && [ -x "${SHIPWRIGHT_BIN}" ]; then
  exec "${SHIPWRIGHT_BIN}" "$@"
fi
if [ -n "${LOOM_BIN:-}" ] && [ -x "${LOOM_BIN}" ]; then
  exec "${LOOM_BIN}" "$@"
fi
if [ -x "../shipwright" ]; then
  exec ../shipwright "$@"
fi
if [ -x "../shipwright.exe" ]; then
  exec ../shipwright.exe "$@"
fi
if [ -x "../loom" ]; then
  exec ../loom "$@"
fi
if command -v shipwright >/dev/null 2>&1; then
  exec shipwright "$@"
fi
if command -v loom >/dev/null 2>&1; then
  exec loom "$@"
fi
if command -v harness >/dev/null 2>&1; then
  exec harness "$@"
fi
echo "Shipwright CLI not found. Install shipwright globally, set SHIPWRIGHT_BIN, or keep the binary one directory above this project." >&2
exit 127
`
}

func opencodeShipwrightWrapperCMD() string {
	return `@echo off
if not "%SHIPWRIGHT_BIN%"=="" if exist "%SHIPWRIGHT_BIN%" "%SHIPWRIGHT_BIN%" %* & exit /b %errorlevel%
if not "%LOOM_BIN%"=="" if exist "%LOOM_BIN%" "%LOOM_BIN%" %* & exit /b %errorlevel%
if exist ..\shipwright.exe ..\shipwright.exe %* & exit /b %errorlevel%
if exist ..\shipwright ..\shipwright %* & exit /b %errorlevel%
if exist ..\loom.exe ..\loom.exe %* & exit /b %errorlevel%
if exist ..\loom ..\loom %* & exit /b %errorlevel%
where shipwright >nul 2>nul
if %errorlevel%==0 shipwright %* & exit /b %errorlevel%
where loom >nul 2>nul
if %errorlevel%==0 loom %* & exit /b %errorlevel%
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
	switch name {
	case "ui-ux-designer":
		return map[string]any{
			"read":          true,
			"write":         true,
			"edit":          true,
			"bash":          false,
			"task":          false,
			"open-pencil_*": true,
		}
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

func opencodeAgentsMD() string {
	var sb strings.Builder
	sb.WriteString(genericAgentsMD())
	sb.WriteString("\n## OpenCode integration\n\n")
	sb.WriteString("This project has been bootstrapped by Shipwright for OpenCode.\n\n")
	sb.WriteString("OpenCode-supported files generated by Shipwright:\n\n")
	sb.WriteString("- `.opencode/opencode.json` — project OpenCode config colocated with executor assets.\n")
	sb.WriteString("- `.opencode/agents/*.md` — role-specific OpenCode agents.\n")
	sb.WriteString("- `.opencode/commands/*.md` — slash commands for Shipwright workflows.\n")
	sb.WriteString("- `.opencode/skills/*/SKILL.md` — reusable role instructions derived from Shipwright agents.\n")
	sb.WriteString("- `.harness/project-profile.md` — calibrated stack, commands, tests, structure, and TDD capability.\n")
	sb.WriteString("- `.harness/tdd-policy.md` — strict/suggested/none TDD policy and evidence rules.\n\n")
	sb.WriteString("Use OpenCode as executor, not as lifecycle authority. Shipwright remains the source of truth.\n\n")
	sb.WriteString("## Shipwright Orchestrator Autopilot\n\n")
	sb.WriteString("You are the default OpenCode agent for this project (`shipwright-orchestrator`). The user should be able to type a product request in natural language and have Shipwright begin the lifecycle automatically.\n\n")
	sb.WriteString("Always use the project-local Shipwright CLI wrapper:\n\n")
	sb.WriteString("- macOS/Linux: `.harness/bin/shipwright <command>`\n")
	sb.WriteString("- Windows: `.harness/bin/shipwright.cmd <command>`\n\n")
	sb.WriteString("Do not assume a global binary is installed; use the project-local wrapper.\n\n")
	sb.WriteString("### On every user message\n\n")
	sb.WriteString("1. Read `.harness/project-profile.md`, `.harness/tdd-policy.md`, `.harness/skill-registry.md`, and `.harness/skill-digests.md` if present; use detected commands, TDD capability, strict TDD policy, and role-specific skill digests instead of inventing stack assumptions.\n")
	sb.WriteString("2. Run `.harness/bin/shipwright status` to read the current lifecycle phase.\n")
	sb.WriteString("3. If phase is `INTAKE` and the user message is a product/build request, run `.harness/bin/shipwright start \"<verbatim user request>\"`.\n")
	sb.WriteString("4. Run `.harness/bin/shipwright agents active` to identify the active role.\n")
	sb.WriteString("5. Delegate role work using the matching Shipwright subagent (`product-owner`, `technical-lead`, `ui-ux-designer`, etc.). Use OpenCode's task/subagent capability when available; otherwise follow that role's `.opencode/skills/<role>/SKILL.md` exactly.\n")
	sb.WriteString("6. Return the role output to the user conversationally. If the active role needs user answers, ask the questions in chat; do not ask the user to manually fill files.\n")
	sb.WriteString("7. After the user answers Product Owner questions, have `product-owner` update the required product artifacts. Do not ask the user to edit files manually.\n")
	sb.WriteString("8. For non-approval internal transitions, run `.harness/bin/shipwright next` yourself when required artifacts/evidence exist. Do not ask the user to run `next`.\n")
	sb.WriteString("9. If Product Owner has produced context, assumptions, open questions, and scope, advance through safe internal phases until `SCOPE_REVIEW`, then present the scope to the user.\n")
	sb.WriteString("10. At approval gates, explain the artifact and ask the user to approve or request changes. Never self-approve. If the user says approval intent like `aprobar scope`, run the matching `.harness/bin/shipwright approve <gate>` yourself.\n")
	sb.WriteString("11. Before moving from IMPLEMENTATION to INTEGRATION, run `.harness/bin/shipwright tdd status`; if mode is `strict`, ensure frontend/backend progress or `reports/tdd-report.md` contains executed test evidence.\n\n")
	sb.WriteString("### Approval UX\n\n")
	sb.WriteString("Do not tell the user to execute raw approval commands unless they explicitly ask for CLI instructions. Instead, ask in natural language. Examples:\n\n")
	sb.WriteString("- `¿Aprobás este alcance o querés cambios?`\n")
	sb.WriteString("- If user says `aprobado`, `aprobar scope`, `sí, aprobar`, or equivalent, run `.harness/bin/shipwright approve scope`.\n")
	sb.WriteString("- If user requests changes, run `.harness/bin/shipwright request-change \"<reason>\"` and route back to the active role.\n")
	sb.WriteString("After approving a gate, continue safe internal transitions yourself until the next role needs user input or another approval gate appears.\n\n")
	sb.WriteString("### DISCOVERY behavior\n\n")
	sb.WriteString("When active agent is `product-owner`, the Product Owner must ask 3-7 concrete discovery questions in the chat before writing final context/scope if critical information is missing. This should feel like a real PO interview, not a file checklist.\n")
	sb.WriteString("If the user's first request is already specific enough, still ask only the minimum useful questions needed to avoid wrong scope; do not ask technical stack questions in discovery unless the user brought them up.\n")
	sb.WriteString("For a billing/invoicing MVP, ask about users, invoice lifecycle, product/customer fields, draft vs issued states, permissions, reporting, and explicit out-of-scope legal/tax requirements.\n\n")
	sb.WriteString("### OpenPencil MCP behavior\n\n")
	sb.WriteString("When the active role is `ui-ux-designer` and OpenPencil is enabled/configured, treat `installed_no_active_canvas` as **unverified**, not as failure. Shipwright cannot validate an active OpenPencil canvas from plain CLI detection; only the MCP client can.\n")
	sb.WriteString("Before falling back to doc-only design, the UI/UX Designer must try the actual OpenCode MCP tools for the `open-pencil` server. OpenCode registers MCP tools with the server name as prefix, so expected tool names should appear as `open-pencil_*`.\n")
	sb.WriteString("If both `pencil` and `open-pencil` MCP servers are connected, do **not** use the `pencil` server for Shipwright. In some environments `pencil` belongs to another desktop host (for example Antigravity) and can fail even when `open-pencil` is healthy.\n")
	sb.WriteString("The validation order is: use `open-pencil_get_editor_state` from the `open-pencil` server; if it succeeds, continue with OpenPencil even if `.harness/bin/shipwright status` says `installed_no_active_canvas`; if no `open-pencil_*` tools are visible or the call fails, report MCP not connected and only then use doc-only fallback.\n\n")
	sb.WriteString("### Safety\n\n")
	sb.WriteString("- Never skip Shipwright gates.\n")
	sb.WriteString("- Never implement code during Product Owner or planning phases.\n")
	sb.WriteString("- Never delete mocks or evidence to pass a gate.\n")
	sb.WriteString("- If a command fails, report the exact failure and safest next action.\n")
	return sb.String()
}

func opencodeAgentMarkdown(skill AgentSkill) string {
	desc := extractSkillDescription(skill.Content)
	permission := opencodePermissionForAgent(skill.Name)
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("description: %s\n", yamlQuote(desc)))
	sb.WriteString("mode: subagent\n")
	sb.WriteString("permission:\n")
	sb.WriteString(fmt.Sprintf("  edit: %s\n", permission.Edit))
	sb.WriteString(fmt.Sprintf("  bash: %s\n", permission.Bash))
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("# Shipwright %s Agent\n\n", skill.Name))
	sb.WriteString("You are executing inside a Shipwright-managed project. Shipwright controls lifecycle, phase gates, approvals, contracts, and evidence.\n\n")
	sb.WriteString("## Before acting\n\n")
	sb.WriteString("1. Read or request `.harness/project-profile.md`, `.harness/tdd-policy.md`, `.harness/skill-registry.md`, and `.harness/skill-digests.md` to understand detected stack, commands, TDD mode, strict evidence policy, and available reusable skills.\n")
	sb.WriteString("2. Run or request `.harness/bin/shipwright status`.\n")
	sb.WriteString("3. Run or request `.harness/bin/shipwright agents active`.\n")
	sb.WriteString(fmt.Sprintf("4. Load and follow `.opencode/skills/%s/SKILL.md`.\n", skill.Name))
	sb.WriteString("5. Do not advance phases or approve gates unless the user explicitly asks you to run the matching Shipwright command.\n\n")
	if skill.Name == "ui-ux-designer" {
		sb.WriteString("## OpenPencil MCP validation\n\n")
		sb.WriteString("If `.harness/integrations.json` enables OpenPencil or `.opencode/opencode.json` contains `mcp.open-pencil`, do not treat `installed_no_active_canvas` as terminal. First try the actual OpenCode MCP tools for `open-pencil`.\n\n")
		sb.WriteString("- Preferred tool pattern: `open-pencil_*` (OpenCode registers MCP tools with server-name prefixes).\n")
		sb.WriteString("- If a separate `pencil` MCP server is connected, do not use it for Shipwright OpenPencil work; it may be bound to another desktop host.\n")
		sb.WriteString("- First validation call: `open-pencil_get_editor_state`.\n")
		sb.WriteString("- Only fall back to doc-only mode if no `open-pencil_*` MCP tool is available or the editor-state call fails.\n\n")
	}
	sb.WriteString("## Role source\n\n")
	sb.WriteString(fmt.Sprintf("Full role instructions live at `.opencode/skills/%s/SKILL.md` and `.harness/agents/%s.md`.\n", skill.Name, skill.Name))
	sb.WriteString("If those differ, prefer `.harness/agents/` because Shipwright generated it as lifecycle source of truth.\n")
	return sb.String()
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
		{Filename: "shipwright-status.md", Content: opencodeCommandMarkdown("Shipwright status", "Run `.harness/bin/shipwright status`, summarize current phase, active gates, missing artifacts, and the safest next Shipwright command. Do not modify files.", "plan")},
		{Filename: "shipwright-active-agent.md", Content: opencodeCommandMarkdown("Shipwright active agent", "Run `.harness/bin/shipwright agents active`, identify the active Shipwright agent, then read `.harness/bin/shipwright agents run <agent>` before proposing work. Do not modify files unless the active role allows it.", "plan")},
		{Filename: "shipwright-next.md", Content: opencodeCommandMarkdown("Shipwright next gate", "Run `.harness/bin/shipwright status` and determine whether `.harness/bin/shipwright next` is safe. If this is a non-approval internal transition and gates/evidence are satisfied, run `.harness/bin/shipwright next` yourself. If an approval gate is blocking, present the artifact and ask the user to approve or request changes.", "plan")},
		{Filename: "shipwright-doctor.md", Content: opencodeCommandMarkdown("Shipwright doctor", "Run `.harness/bin/shipwright doctor` and summarize blocking errors, warnings, fallbacks, and concrete fixes. Do not edit config unless the user explicitly asks for `.harness/bin/shipwright doctor --fix`.", "plan")},
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
