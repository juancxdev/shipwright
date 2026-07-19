package harness

import (
	"encoding/json"
	"os"
)

const IntegrationsFile = ".harness/integrations.json"

type IntegrationConfig struct {
	Enabled        bool   `json:"enabled"`
	Mode           string `json:"mode"`
	Status         string `json:"status,omitempty"`
	Fallback       string `json:"fallback,omitempty"`
	BinaryPath     string `json:"binary_path,omitempty"`
	AppPath        string `json:"app_path,omitempty"`
	MCPServerPath  string `json:"mcp_server_path,omitempty"`
	MCPCommand     string `json:"mcp_command,omitempty"`
	HealthURL      string `json:"health_url,omitempty"`
	Version        string `json:"version,omitempty"`
	LastDetectedAt string `json:"last_detected_at,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

type Integrations struct {
	Platform   PlatformInfo      `json:"platform,omitempty"`
	Engram     IntegrationConfig `json:"engram"`
	OpenPencil IntegrationConfig `json:"openpencil"`
}

func DefaultIntegrations() *Integrations {
	return &Integrations{
		Engram: IntegrationConfig{
			Enabled:   false,
			Mode:      "mcp",
			Status:    "not_configured",
			Fallback:  "progress/decisions.md",
			HealthURL: "http://localhost:7437/health",
		},
		OpenPencil: IntegrationConfig{
			Enabled:  false,
			Mode:     "mcp",
			Status:   "not_configured",
			Fallback: "design-doc-only",
		},
	}
}

func (i *Integrations) ApplyPortableConfig(cfg *PortableConfig) {
	if i == nil || cfg == nil {
		return
	}
	cfg.Normalize()

	i.Engram.Mode = cfg.Integrations.Engram.Mode
	i.Engram.BinaryPath = cfg.Integrations.Engram.BinaryPath
	i.Engram.HealthURL = cfg.Integrations.Engram.HealthURL
	i.Engram.Fallback = cfg.Integrations.Engram.Fallback

	i.OpenPencil.Mode = cfg.Integrations.OpenPencil.Mode
	i.OpenPencil.AppPath = cfg.Integrations.OpenPencil.AppPath
	i.OpenPencil.MCPServerPath = cfg.Integrations.OpenPencil.MCPServerPath
	i.OpenPencil.MCPCommand = cfg.Integrations.OpenPencil.MCPCommand
	i.OpenPencil.Fallback = cfg.Integrations.OpenPencil.Fallback
}

func (i *Integrations) ApplyDetection(engram DetectionResult, openpencil DetectionResult) {
	i.Platform = engram.Platform
	if i.Platform.OS == "" {
		i.Platform = openpencil.Platform
	}

	i.Engram.Status = engram.Status
	i.Engram.Reason = engram.Reason
	i.Engram.BinaryPath = engram.Path
	i.Engram.Version = engram.Version
	i.Engram.Fallback = engram.Fallback
	i.Engram.LastDetectedAt = NowISO()
	if i.Engram.HealthURL == "" {
		i.Engram.HealthURL = "http://localhost:7437/health"
	}

	i.OpenPencil.Status = openpencil.Status
	i.OpenPencil.Reason = openpencil.Reason
	if openpencil.Path != "" {
		i.OpenPencil.AppPath = ""
		i.OpenPencil.MCPServerPath = ""
		i.OpenPencil.MCPCommand = ""
	}
	switch openpencil.PathKind {
	case DetectionPathMCPServer:
		i.OpenPencil.MCPServerPath = openpencil.Path
	case DetectionPathBinary:
		i.OpenPencil.MCPCommand = openpencil.Path
	case DetectionPathApp:
		i.OpenPencil.AppPath = openpencil.Path
	}
	i.OpenPencil.Fallback = openpencil.Fallback
	i.OpenPencil.LastDetectedAt = NowISO()
}

func LoadIntegrations() (*Integrations, error) {
	data, err := os.ReadFile(IntegrationsFile)
	if err != nil {
		return nil, err
	}
	var i Integrations
	if err := json.Unmarshal(data, &i); err != nil {
		return nil, err
	}
	return &i, nil
}

func (i *Integrations) Save() error {
	data, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(IntegrationsFile, data, 0644)
}

func (i *Integrations) EnableEngram() {
	i.Engram.Enabled = true
	i.Engram.Status = "enabled_via_cli"
}

func (i *Integrations) DisableEngram() {
	i.Engram.Enabled = false
	i.Engram.Status = "disabled_via_cli"
}

func (i *Integrations) IsEngramEnabled() bool {
	return i.Engram.Enabled
}

func (i *Integrations) EnableOpenPencil() {
	i.OpenPencil.Enabled = true
	i.OpenPencil.Status = "enabled_via_cli"
}

func (i *Integrations) DisableOpenPencil() {
	i.OpenPencil.Enabled = false
	i.OpenPencil.Status = "disabled_via_cli"
}

func (i *Integrations) IsOpenPencilEnabled() bool {
	return i.OpenPencil.Enabled
}

func (i *Integrations) SetOpenPencilStatus(status string) {
	i.OpenPencil.Status = status
}
