package application

import (
	"fmt"
	"net/url"
	"strings"

	agents "shipwright/internal/agents/application"
)

const (
	ConfigIssueSeverityError   = "ERROR"
	ConfigIssueSeverityWarning = "WARNING"
)

const (
	ConfigModeMCP      = "mcp"
	ConfigModeSDK      = "sdk"
	ConfigModeDisabled = "disabled"
)

type ConfigValidationIssue struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Action   string `json:"action,omitempty"`
}

func ValidatePortableConfig(cfg *PortableConfig) []ConfigValidationIssue {
	if cfg == nil {
		return []ConfigValidationIssue{{
			ID:       "config.nil",
			Path:     PortableConfigFile,
			Severity: ConfigIssueSeverityError,
			Message:  "portable config is nil",
			Action:   "Run 'shipwright doctor --fix' to recreate .harness/config.json.",
		}}
	}

	var issues []ConfigValidationIssue
	issues = append(issues, validateConfigCore(cfg)...)
	issues = append(issues, validateIntegrationsConfig("integrations", cfg.Integrations, true)...)
	issues = append(issues, validateExecutorsConfig(cfg.Executors)...)
	for osName, override := range cfg.PlatformOverrides {
		path := fmt.Sprintf("platform_overrides.%s", osName)
		if strings.TrimSpace(osName) == "" {
			issues = append(issues, ConfigValidationIssue{
				ID:       "config.platform_override.empty_key",
				Path:     "platform_overrides",
				Severity: ConfigIssueSeverityWarning,
				Message:  "platform override has an empty OS key",
				Action:   "Remove the empty platform override key.",
			})
		}
		issues = append(issues, validateIntegrationsConfig(path, override, false)...)
	}
	return issues
}

func validateConfigCore(cfg *PortableConfig) []ConfigValidationIssue {
	var issues []ConfigValidationIssue
	if strings.TrimSpace(cfg.Version) != PortableConfigVersion {
		issues = append(issues, ConfigValidationIssue{
			ID:       "config.version.unsupported",
			Path:     "version",
			Severity: ConfigIssueSeverityError,
			Message:  fmt.Sprintf("unsupported config version %q", cfg.Version),
			Action:   fmt.Sprintf("Set version to %q or regenerate config with 'shipwright doctor --fix'.", PortableConfigVersion),
		})
	}
	if strings.TrimSpace(cfg.ArtifactRoot) == "" {
		issues = append(issues, ConfigValidationIssue{
			ID:       "config.artifact_root.empty",
			Path:     "artifact_root",
			Severity: ConfigIssueSeverityError,
			Message:  "artifact_root cannot be empty",
			Action:   "Set artifact_root to '.'.",
		})
	}
	if cfg.Health.TimeoutMillis <= 0 {
		issues = append(issues, ConfigValidationIssue{
			ID:       "config.health.timeout.invalid",
			Path:     "health.timeout_millis",
			Severity: ConfigIssueSeverityError,
			Message:  "health timeout must be greater than 0",
			Action:   fmt.Sprintf("Set health.timeout_millis to %d or another positive value.", DefaultHealthTimeoutMillis),
		})
	} else if cfg.Health.TimeoutMillis > 30000 {
		issues = append(issues, ConfigValidationIssue{
			ID:       "config.health.timeout.too_high",
			Path:     "health.timeout_millis",
			Severity: ConfigIssueSeverityWarning,
			Message:  "health timeout is very high and can make doctor feel stuck",
			Action:   "Consider using 1500-5000 ms.",
		})
	}
	return issues
}

func validateIntegrationsConfig(prefix string, cfg PortableIntegrationsConfig, requireFallbacks bool) []ConfigValidationIssue {
	var issues []ConfigValidationIssue
	issues = append(issues, validateMode(prefix+".engram.mode", cfg.Engram.Mode)...)
	issues = append(issues, validateMode(prefix+".stitch.mode", cfg.Stitch.Mode)...)
	issues = append(issues, validateMode(prefix+".opendesign.mode", cfg.OpenDesign.Mode)...)
	issues = append(issues, validateMode(prefix+".openpencil.mode", cfg.OpenPencil.Mode)...)
	if requireFallbacks && strings.TrimSpace(cfg.Engram.Fallback) == "" {
		issues = append(issues, ConfigValidationIssue{
			ID:       "config.engram.fallback.empty",
			Path:     prefix + ".engram.fallback",
			Severity: ConfigIssueSeverityError,
			Message:  "Engram fallback cannot be empty",
			Action:   "Set fallback to .harness/artifacts/progress/decisions.md.",
		})
	}
	if requireFallbacks && strings.TrimSpace(cfg.Stitch.Fallback) == "" {
		issues = append(issues, ConfigValidationIssue{
			ID:       "config.stitch.fallback.empty",
			Path:     prefix + ".stitch.fallback",
			Severity: ConfigIssueSeverityError,
			Message:  "Stitch fallback cannot be empty",
			Action:   "Set fallback to design-doc-only.",
		})
	}
	if requireFallbacks && strings.TrimSpace(cfg.OpenDesign.Fallback) == "" {
		issues = append(issues, ConfigValidationIssue{
			ID:       "config.opendesign.fallback.empty",
			Path:     prefix + ".opendesign.fallback",
			Severity: ConfigIssueSeverityError,
			Message:  "OpenDesign fallback cannot be empty",
			Action:   "Set fallback to design-doc-only.",
		})
	}
	if requireFallbacks && strings.TrimSpace(cfg.OpenPencil.Fallback) == "" {
		issues = append(issues, ConfigValidationIssue{
			ID:       "config.openpencil.fallback.empty",
			Path:     prefix + ".openpencil.fallback",
			Severity: ConfigIssueSeverityError,
			Message:  "OpenPencil fallback cannot be empty",
			Action:   "Set fallback to design-doc-only.",
		})
	}
	if strings.TrimSpace(cfg.Engram.HealthURL) != "" {
		if issue := ValidateHTTPURL(prefix+".engram.health_url", cfg.Engram.HealthURL); issue != nil {
			issues = append(issues, *issue)
		}
	}
	if strings.TrimSpace(cfg.Stitch.MCPURL) != "" {
		if issue := ValidateHTTPURL(prefix+".stitch.mcp_url", cfg.Stitch.MCPURL); issue != nil {
			issues = append(issues, *issue)
		}
	}
	issues = append(issues, validatePathShape(prefix+".engram.binary_path", cfg.Engram.BinaryPath)...)
	issues = append(issues, validateCommandShape(prefix+".opendesign.mcp_command", cfg.OpenDesign.MCPCommand)...)
	for index, arg := range cfg.OpenDesign.MCPArgs {
		issues = append(issues, validateCommandShape(fmt.Sprintf("%s.opendesign.mcp_args.%d", prefix, index), arg)...)
	}
	if strings.TrimSpace(cfg.OpenDesign.DaemonURL) != "" {
		if issue := ValidateHTTPURL(prefix+".opendesign.daemon_url", cfg.OpenDesign.DaemonURL); issue != nil {
			issues = append(issues, *issue)
		}
	}
	issues = append(issues, validatePathShape(prefix+".opendesign.data_dir", cfg.OpenDesign.DataDir)...)
	issues = append(issues, validatePathShape(prefix+".opendesign.ipc_path", cfg.OpenDesign.IPCPath)...)
	issues = append(issues, validatePathShape(prefix+".openpencil.app_path", cfg.OpenPencil.AppPath)...)
	issues = append(issues, validatePathShape(prefix+".openpencil.mcp_server_path", cfg.OpenPencil.MCPServerPath)...)
	issues = append(issues, validateCommandShape(prefix+".openpencil.mcp_command", cfg.OpenPencil.MCPCommand)...)
	return issues
}

func validateExecutorsConfig(cfg PortableExecutorsConfig) []ConfigValidationIssue {
	var issues []ConfigValidationIssue
	issues = append(issues, validateOptionalModel("executors.opencode.default_model", cfg.OpenCode.DefaultModel)...)
	issues = append(issues, validateOptionalModel("executors.opencode.balanced_model", cfg.OpenCode.BalancedModel)...)
	issues = append(issues, validateOptionalModel("executors.opencode.reasoning_model", cfg.OpenCode.ReasoningModel)...)
	issues = append(issues, validateOptionalModel("executors.opencode.fast_model", cfg.OpenCode.FastModel)...)
	knownAgents := knownOpenCodeAgents()
	for agent, model := range cfg.OpenCode.AgentModels {
		agentPath := fmt.Sprintf("executors.opencode.agent_models.%s", agent)
		if strings.TrimSpace(agent) == "" {
			issues = append(issues, ConfigValidationIssue{
				ID:       "config.opencode.agent_model.empty_agent",
				Path:     "executors.opencode.agent_models",
				Severity: ConfigIssueSeverityError,
				Message:  "OpenCode agent model override has an empty agent name",
				Action:   "Remove the empty key or replace it with a known Shipwright agent name.",
			})
			continue
		}
		if strings.TrimSpace(model) == "" {
			issues = append(issues, ConfigValidationIssue{
				ID:       "config.opencode.agent_model.empty_model",
				Path:     agentPath,
				Severity: ConfigIssueSeverityError,
				Message:  fmt.Sprintf("OpenCode model for agent %q cannot be empty", agent),
				Action:   "Set a provider/model id like 'opencode-go/deepseek-v4-flash'.",
			})
		}
		if !knownAgents[agent] {
			issues = append(issues, ConfigValidationIssue{
				ID:       "config.opencode.agent_model.unknown_agent",
				Path:     agentPath,
				Severity: ConfigIssueSeverityWarning,
				Message:  fmt.Sprintf("unknown OpenCode agent %q", agent),
				Action:   "Use one of Shipwright's generated agents or remove this override.",
			})
		}
	}
	return issues
}

func validateOptionalModel(path, value string) []ConfigValidationIssue {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if strings.ContainsAny(value, "\n\r\t") {
		return []ConfigValidationIssue{{
			ID:       "config.opencode.model.invalid_chars",
			Path:     path,
			Severity: ConfigIssueSeverityError,
			Message:  "OpenCode model id contains control characters",
			Action:   "Use a clean provider/model id.",
		}}
	}
	return nil
}

func validateMode(path, mode string) []ConfigValidationIssue {
	mode = strings.TrimSpace(mode)
	if mode == "" || mode == ConfigModeMCP || mode == ConfigModeSDK || mode == ConfigModeDisabled {
		return nil
	}
	return []ConfigValidationIssue{{
		ID:       "config.mode.unsupported",
		Path:     path,
		Severity: ConfigIssueSeverityError,
		Message:  fmt.Sprintf("unsupported mode %q", mode),
		Action:   "Use 'mcp', 'sdk', or 'disabled'.",
	}}
}

func ValidateHTTPURL(path, raw string) *ConfigValidationIssue {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return &ConfigValidationIssue{
			ID:       "config.url.invalid",
			Path:     path,
			Severity: ConfigIssueSeverityError,
			Message:  fmt.Sprintf("invalid URL %q", raw),
			Action:   "Use an absolute http:// or https:// URL.",
		}
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return &ConfigValidationIssue{
			ID:       "config.url.unsupported_scheme",
			Path:     path,
			Severity: ConfigIssueSeverityError,
			Message:  fmt.Sprintf("unsupported URL scheme %q", parsed.Scheme),
			Action:   "Use http:// or https://.",
		}
	}
	return nil
}

func validatePathShape(path, value string) []ConfigValidationIssue {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if strings.ContainsAny(value, "\n\r\t") {
		return []ConfigValidationIssue{{
			ID:       "config.path.invalid_chars",
			Path:     path,
			Severity: ConfigIssueSeverityError,
			Message:  "path contains control characters",
			Action:   "Replace it with a clean absolute path or remove the value.",
		}}
	}
	if looksRelativePath(value) {
		return []ConfigValidationIssue{{
			ID:       "config.path.relative",
			Path:     path,
			Severity: ConfigIssueSeverityWarning,
			Message:  fmt.Sprintf("path %q looks relative", value),
			Action:   "Prefer absolute paths for portable config, or use environment variables per machine.",
		}}
	}
	return nil
}

func validateCommandShape(path, value string) []ConfigValidationIssue {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if strings.ContainsAny(value, "\n\r\t") {
		return []ConfigValidationIssue{{
			ID:       "config.command.invalid_chars",
			Path:     path,
			Severity: ConfigIssueSeverityError,
			Message:  "command contains control characters",
			Action:   "Use a clean command like 'openpencil-mcp' or remove the value.",
		}}
	}
	return nil
}

func looksRelativePath(value string) bool {
	if strings.HasPrefix(value, "/") || strings.HasPrefix(value, "~") {
		return false
	}
	if len(value) >= 3 && ((value[1] == ':' && (value[2] == '\\' || value[2] == '/')) || strings.HasPrefix(value, "\\\\")) {
		return false
	}
	return true
}

func CountConfigIssueSeverities(issues []ConfigValidationIssue) (errorsCount int, warningsCount int) {
	for _, issue := range issues {
		switch issue.Severity {
		case ConfigIssueSeverityError:
			errorsCount++
		case ConfigIssueSeverityWarning:
			warningsCount++
		}
	}
	return errorsCount, warningsCount
}

func knownOpenCodeAgents() map[string]bool {
	known := map[string]bool{"shipwright-orchestrator": true}
	for _, agent := range agents.AllAgents() {
		known[agent.Name] = true
	}
	return known
}
