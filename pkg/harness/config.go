package harness

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
)

const PortableConfigFile = ".harness/config.json"
const PortableConfigVersion = "1"

type PortableConfig struct {
	Version           string                                `json:"version"`
	ArtifactRoot      string                                `json:"artifact_root"`
	Health            PortableHealthConfig                  `json:"health"`
	Integrations      PortableIntegrationsConfig            `json:"integrations"`
	Executors         PortableExecutorsConfig               `json:"executors"`
	PlatformOverrides map[string]PortableIntegrationsConfig `json:"platform_overrides,omitempty"`
}

type PortableHealthConfig struct {
	TimeoutMillis int `json:"timeout_millis"`
}

type PortableIntegrationsConfig struct {
	Engram     PortableEngramConfig     `json:"engram"`
	Stitch     PortableStitchConfig     `json:"stitch"`
	OpenDesign PortableOpenDesignConfig `json:"opendesign"`
	OpenPencil PortableOpenPencilConfig `json:"openpencil"`
}

type PortableExecutorsConfig struct {
	OpenCode PortableOpenCodeExecutorConfig `json:"opencode"`
}

type PortableOpenCodeExecutorConfig struct {
	DefaultModel   string            `json:"default_model,omitempty"`
	ReasoningModel string            `json:"reasoning_model,omitempty"`
	FastModel      string            `json:"fast_model,omitempty"`
	AgentModels    map[string]string `json:"agent_models,omitempty"`
}

type PortableEngramConfig struct {
	Mode       string `json:"mode"`
	BinaryPath string `json:"binary_path,omitempty"`
	HealthURL  string `json:"health_url,omitempty"`
	Fallback   string `json:"fallback"`
}

type PortableStitchConfig struct {
	Mode           string `json:"mode"`
	APIKeyEnv      string `json:"api_key_env,omitempty"`
	AccessTokenEnv string `json:"access_token_env,omitempty"`
	ProjectEnv     string `json:"project_env,omitempty"`
	MCPURL         string `json:"mcp_url,omitempty"`
	Fallback       string `json:"fallback"`
}

type PortableOpenDesignConfig struct {
	Mode       string   `json:"mode"`
	MCPCommand string   `json:"mcp_command,omitempty"`
	MCPArgs    []string `json:"mcp_args,omitempty"`
	DataDir    string   `json:"data_dir,omitempty"`
	IPCPath    string   `json:"ipc_path,omitempty"`
	Fallback   string   `json:"fallback"`
}

type PortableOpenPencilConfig struct {
	Mode          string `json:"mode"`
	AppPath       string `json:"app_path,omitempty"`
	MCPServerPath string `json:"mcp_server_path,omitempty"`
	MCPCommand    string `json:"mcp_command,omitempty"`
	Fallback      string `json:"fallback"`
}

func DefaultPortableConfig() *PortableConfig {
	return &PortableConfig{
		Version:      PortableConfigVersion,
		ArtifactRoot: ".",
		Health: PortableHealthConfig{
			TimeoutMillis: DefaultHealthTimeoutMillis,
		},
		Integrations: PortableIntegrationsConfig{
			Engram: PortableEngramConfig{
				Mode:      "mcp",
				HealthURL: "http://localhost:7437/health",
				Fallback:  DecisionsFile,
			},
			Stitch: PortableStitchConfig{
				Mode:           "sdk",
				APIKeyEnv:      "STITCH_API_KEY",
				AccessTokenEnv: "STITCH_ACCESS_TOKEN",
				ProjectEnv:     "GOOGLE_CLOUD_PROJECT",
				MCPURL:         DefaultStitchMCPEndpoint,
				Fallback:       "design-doc-only",
			},
			OpenDesign: PortableOpenDesignConfig{
				Mode:     "mcp",
				Fallback: "design-doc-only",
			},
			OpenPencil: PortableOpenPencilConfig{
				Mode:     "mcp",
				Fallback: "design-doc-only",
			},
		},
		Executors: PortableExecutorsConfig{
			OpenCode: DefaultOpenCodeExecutorConfig(),
		},
		PlatformOverrides: map[string]PortableIntegrationsConfig{},
	}
}

func LoadPortableConfigRaw() (*PortableConfig, error) {
	data, err := os.ReadFile(PortableConfigFile)
	if err != nil {
		return nil, err
	}
	var cfg PortableConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadPortableConfig() (*PortableConfig, error) {
	cfg, err := LoadPortableConfigRaw()
	if err != nil {
		return nil, err
	}
	cfg.Normalize()
	return cfg, nil
}

func LoadEffectivePortableConfig(probe SystemProbe) (*PortableConfig, error) {
	if probe == nil {
		probe = RealSystemProbe{}
	}

	cfg := DefaultPortableConfig()
	loaded, err := LoadPortableConfig()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	} else {
		cfg.Merge(loaded.Integrations)
		cfg.MergeExecutors(loaded.Executors)
		cfg.PlatformOverrides = loaded.PlatformOverrides
		if loaded.Version != "" {
			cfg.Version = loaded.Version
		}
		if loaded.ArtifactRoot != "" {
			cfg.ArtifactRoot = loaded.ArtifactRoot
		}
		if loaded.Health.TimeoutMillis > 0 {
			cfg.Health.TimeoutMillis = loaded.Health.TimeoutMillis
		}
	}

	platform := DetectPlatform(probe)
	if cfg.PlatformOverrides != nil {
		if override, ok := cfg.PlatformOverrides[platform.OS]; ok {
			cfg.Merge(override)
		}
	}

	cfg.ApplyEnv(probe)
	cfg.Normalize()
	return cfg, nil
}

func (cfg *PortableConfig) Save() error {
	if cfg == nil {
		cfg = DefaultPortableConfig()
	}
	cfg.Normalize()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return WriteFile(PortableConfigFile, string(data))
}

func (cfg *PortableConfig) Normalize() {
	if cfg.Version == "" {
		cfg.Version = PortableConfigVersion
	}
	if cfg.ArtifactRoot == "" {
		cfg.ArtifactRoot = "."
	}
	if cfg.Health.TimeoutMillis <= 0 {
		cfg.Health.TimeoutMillis = DefaultHealthTimeoutMillis
	}
	if cfg.Integrations.Engram.Mode == "" {
		cfg.Integrations.Engram.Mode = "mcp"
	}
	if cfg.Integrations.Engram.HealthURL == "" {
		cfg.Integrations.Engram.HealthURL = "http://localhost:7437/health"
	}
	if cfg.Integrations.Engram.Fallback == "" {
		cfg.Integrations.Engram.Fallback = DecisionsFile
	}
	if cfg.Integrations.Stitch.Mode == "" {
		cfg.Integrations.Stitch.Mode = "sdk"
	}
	if cfg.Integrations.Stitch.APIKeyEnv == "" {
		cfg.Integrations.Stitch.APIKeyEnv = "STITCH_API_KEY"
	}
	if cfg.Integrations.Stitch.AccessTokenEnv == "" {
		cfg.Integrations.Stitch.AccessTokenEnv = "STITCH_ACCESS_TOKEN"
	}
	if cfg.Integrations.Stitch.ProjectEnv == "" {
		cfg.Integrations.Stitch.ProjectEnv = "GOOGLE_CLOUD_PROJECT"
	}
	if cfg.Integrations.Stitch.MCPURL == "" {
		cfg.Integrations.Stitch.MCPURL = DefaultStitchMCPEndpoint
	}
	if cfg.Integrations.Stitch.Fallback == "" {
		cfg.Integrations.Stitch.Fallback = "design-doc-only"
	}
	if cfg.Integrations.OpenDesign.Mode == "" {
		cfg.Integrations.OpenDesign.Mode = "mcp"
	}
	if cfg.Integrations.OpenDesign.Fallback == "" {
		cfg.Integrations.OpenDesign.Fallback = "design-doc-only"
	}
	if cfg.Integrations.OpenPencil.Mode == "" {
		cfg.Integrations.OpenPencil.Mode = "mcp"
	}
	if cfg.Integrations.OpenPencil.Fallback == "" {
		cfg.Integrations.OpenPencil.Fallback = "design-doc-only"
	}
	cfg.Executors.OpenCode.Normalize()
	if cfg.PlatformOverrides == nil {
		cfg.PlatformOverrides = map[string]PortableIntegrationsConfig{}
	}
}

func DefaultOpenCodeExecutorConfig() PortableOpenCodeExecutorConfig {
	return PortableOpenCodeExecutorConfig{
		DefaultModel:   "anthropic/claude-sonnet-4-20250514",
		ReasoningModel: "anthropic/claude-sonnet-4-20250514",
		FastModel:      "anthropic/claude-haiku-4-20250514",
		AgentModels:    map[string]string{},
	}
}

func (cfg *PortableOpenCodeExecutorConfig) Normalize() {
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = "anthropic/claude-sonnet-4-20250514"
	}
	if cfg.ReasoningModel == "" {
		cfg.ReasoningModel = cfg.DefaultModel
	}
	if cfg.FastModel == "" {
		cfg.FastModel = cfg.DefaultModel
	}
	if cfg.AgentModels == nil {
		cfg.AgentModels = map[string]string{}
	}
}

func (cfg *PortableConfig) Merge(override PortableIntegrationsConfig) {
	if override.Engram.Mode != "" {
		cfg.Integrations.Engram.Mode = override.Engram.Mode
	}
	if override.Engram.BinaryPath != "" {
		cfg.Integrations.Engram.BinaryPath = override.Engram.BinaryPath
	}
	if override.Engram.HealthURL != "" {
		cfg.Integrations.Engram.HealthURL = override.Engram.HealthURL
	}
	if override.Engram.Fallback != "" {
		cfg.Integrations.Engram.Fallback = override.Engram.Fallback
	}

	if override.Stitch.Mode != "" {
		cfg.Integrations.Stitch.Mode = override.Stitch.Mode
	}
	if override.Stitch.APIKeyEnv != "" {
		cfg.Integrations.Stitch.APIKeyEnv = override.Stitch.APIKeyEnv
	}
	if override.Stitch.AccessTokenEnv != "" {
		cfg.Integrations.Stitch.AccessTokenEnv = override.Stitch.AccessTokenEnv
	}
	if override.Stitch.ProjectEnv != "" {
		cfg.Integrations.Stitch.ProjectEnv = override.Stitch.ProjectEnv
	}
	if override.Stitch.MCPURL != "" {
		cfg.Integrations.Stitch.MCPURL = override.Stitch.MCPURL
	}
	if override.Stitch.Fallback != "" {
		cfg.Integrations.Stitch.Fallback = override.Stitch.Fallback
	}

	if override.OpenDesign.Mode != "" {
		cfg.Integrations.OpenDesign.Mode = override.OpenDesign.Mode
	}
	if override.OpenDesign.MCPCommand != "" {
		cfg.Integrations.OpenDesign.MCPCommand = override.OpenDesign.MCPCommand
	}
	if len(override.OpenDesign.MCPArgs) > 0 {
		cfg.Integrations.OpenDesign.MCPArgs = append([]string{}, override.OpenDesign.MCPArgs...)
	}
	if override.OpenDesign.DataDir != "" {
		cfg.Integrations.OpenDesign.DataDir = override.OpenDesign.DataDir
	}
	if override.OpenDesign.IPCPath != "" {
		cfg.Integrations.OpenDesign.IPCPath = override.OpenDesign.IPCPath
	}
	if override.OpenDesign.Fallback != "" {
		cfg.Integrations.OpenDesign.Fallback = override.OpenDesign.Fallback
	}

	if override.OpenPencil.Mode != "" {
		cfg.Integrations.OpenPencil.Mode = override.OpenPencil.Mode
	}
	if override.OpenPencil.AppPath != "" {
		cfg.Integrations.OpenPencil.AppPath = override.OpenPencil.AppPath
	}
	if override.OpenPencil.MCPServerPath != "" {
		cfg.Integrations.OpenPencil.MCPServerPath = override.OpenPencil.MCPServerPath
	}
	if override.OpenPencil.MCPCommand != "" {
		cfg.Integrations.OpenPencil.MCPCommand = override.OpenPencil.MCPCommand
	}
	if override.OpenPencil.Fallback != "" {
		cfg.Integrations.OpenPencil.Fallback = override.OpenPencil.Fallback
	}
}

func (cfg *PortableConfig) MergeExecutors(override PortableExecutorsConfig) {
	if override.OpenCode.DefaultModel != "" {
		cfg.Executors.OpenCode.DefaultModel = override.OpenCode.DefaultModel
	}
	if override.OpenCode.ReasoningModel != "" {
		cfg.Executors.OpenCode.ReasoningModel = override.OpenCode.ReasoningModel
	}
	if override.OpenCode.FastModel != "" {
		cfg.Executors.OpenCode.FastModel = override.OpenCode.FastModel
	}
	if len(override.OpenCode.AgentModels) > 0 {
		if cfg.Executors.OpenCode.AgentModels == nil {
			cfg.Executors.OpenCode.AgentModels = map[string]string{}
		}
		for agent, model := range override.OpenCode.AgentModels {
			cfg.Executors.OpenCode.AgentModels[agent] = model
		}
	}
}

func (cfg *PortableConfig) ApplyEnv(probe SystemProbe) {
	if probe == nil {
		probe = RealSystemProbe{}
	}
	if value := trimEnv(probe.Getenv("SHIPWRIGHT_HEALTH_TIMEOUT_MS")); value != "" {
		if millis, err := strconv.Atoi(value); err == nil && millis > 0 {
			cfg.Health.TimeoutMillis = millis
		}
	}
	if value := trimEnv(probe.Getenv("ENGRAM_BINARY")); value != "" {
		cfg.Integrations.Engram.BinaryPath = value
	}
	if value := trimEnv(probe.Getenv("ENGRAM_HEALTH_URL")); value != "" {
		cfg.Integrations.Engram.HealthURL = value
	}
	if value := trimEnv(probe.Getenv("SHIPWRIGHT_STITCH_API_KEY_ENV")); value != "" {
		cfg.Integrations.Stitch.APIKeyEnv = value
	}
	if value := trimEnv(probe.Getenv("SHIPWRIGHT_STITCH_ACCESS_TOKEN_ENV")); value != "" {
		cfg.Integrations.Stitch.AccessTokenEnv = value
	}
	if value := trimEnv(probe.Getenv("SHIPWRIGHT_STITCH_PROJECT_ENV")); value != "" {
		cfg.Integrations.Stitch.ProjectEnv = value
	}
	if value := trimEnv(probe.Getenv("SHIPWRIGHT_STITCH_MCP_URL")); value != "" {
		cfg.Integrations.Stitch.MCPURL = value
	}
	if value := trimEnv(probe.Getenv("OPENDESIGN_MCP_COMMAND")); value != "" {
		cfg.Integrations.OpenDesign.MCPCommand = value
	}
	if value := trimEnv(probe.Getenv("OPENDESIGN_MCP_ARGS")); value != "" {
		cfg.Integrations.OpenDesign.MCPArgs = parseEnvList(value)
	}
	if value := trimEnv(probe.Getenv("OPENDESIGN_DATA_DIR")); value != "" {
		cfg.Integrations.OpenDesign.DataDir = value
	}
	if value := trimEnv(probe.Getenv("OD_DATA_DIR")); value != "" {
		cfg.Integrations.OpenDesign.DataDir = value
	}
	if value := trimEnv(probe.Getenv("OPENDESIGN_SIDECAR_IPC_PATH")); value != "" {
		cfg.Integrations.OpenDesign.IPCPath = value
	}
	if value := trimEnv(probe.Getenv("OD_SIDECAR_IPC_PATH")); value != "" {
		cfg.Integrations.OpenDesign.IPCPath = value
	}
	if value := trimEnv(probe.Getenv("OPENPENCIL_APP_PATH")); value != "" {
		cfg.Integrations.OpenPencil.AppPath = value
	}
	if value := trimEnv(probe.Getenv("OPENPENCIL_MCP_SERVER")); value != "" {
		cfg.Integrations.OpenPencil.MCPServerPath = value
	}
	if value := trimEnv(probe.Getenv("OPENPENCIL_MCP_COMMAND")); value != "" {
		cfg.Integrations.OpenPencil.MCPCommand = value
	}
	if value := trimEnv(probe.Getenv("SHIPWRIGHT_OPENCODE_DEFAULT_MODEL")); value != "" {
		cfg.Executors.OpenCode.DefaultModel = value
	}
	if value := trimEnv(probe.Getenv("SHIPWRIGHT_OPENCODE_REASONING_MODEL")); value != "" {
		cfg.Executors.OpenCode.ReasoningModel = value
	}
	if value := trimEnv(probe.Getenv("SHIPWRIGHT_OPENCODE_FAST_MODEL")); value != "" {
		cfg.Executors.OpenCode.FastModel = value
	}
	if value := trimEnv(probe.Getenv("SHIPWRIGHT_OPENCODE_AGENT_MODELS")); value != "" {
		for agent, model := range ParseAgentModelOverrides(value) {
			if cfg.Executors.OpenCode.AgentModels == nil {
				cfg.Executors.OpenCode.AgentModels = map[string]string{}
			}
			cfg.Executors.OpenCode.AgentModels[agent] = model
		}
	}
}

func trimEnv(value string) string {
	for len(value) > 0 && (value[0] == ' ' || value[0] == '\t' || value[0] == '\n' || value[0] == '\r') {
		value = value[1:]
	}
	for len(value) > 0 {
		last := value[len(value)-1]
		if last != ' ' && last != '\t' && last != '\n' && last != '\r' {
			break
		}
		value = value[:len(value)-1]
	}
	return value
}

func parseEnvList(value string) []string {
	value = trimEnv(value)
	if value == "" {
		return nil
	}
	separator := ","
	if strings.Contains(value, "|") {
		separator = "|"
	}
	parts := strings.Split(value, separator)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = trimEnv(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
