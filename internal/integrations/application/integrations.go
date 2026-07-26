package application

import (
	"encoding/json"
	"os"
	"time"

	config "shipwright/internal/config/application"
	platform "shipwright/internal/platform/application"
)

const IntegrationsFile = ".harness/integrations.json"

type IntegrationConfig struct {
	Enabled        bool     `json:"enabled"`
	Mode           string   `json:"mode"`
	Status         string   `json:"status,omitempty"`
	Fallback       string   `json:"fallback,omitempty"`
	BinaryPath     string   `json:"binary_path,omitempty"`
	AppPath        string   `json:"app_path,omitempty"`
	MCPServerPath  string   `json:"mcp_server_path,omitempty"`
	MCPCommand     string   `json:"mcp_command,omitempty"`
	MCPArgs        []string `json:"mcp_args,omitempty"`
	DataDir        string   `json:"data_dir,omitempty"`
	IPCPath        string   `json:"ipc_path,omitempty"`
	HealthURL      string   `json:"health_url,omitempty"`
	Version        string   `json:"version,omitempty"`
	LastDetectedAt string   `json:"last_detected_at,omitempty"`
	Reason         string   `json:"reason,omitempty"`
}

type Integrations struct {
	Platform   platform.PlatformInfo `json:"platform,omitempty"`
	Engram     IntegrationConfig     `json:"engram"`
	Stitch     IntegrationConfig     `json:"stitch"`
	OpenDesign IntegrationConfig     `json:"opendesign"`
	OpenPencil IntegrationConfig     `json:"openpencil"`
}

func DefaultIntegrations() *Integrations {
	return &Integrations{
		Engram: IntegrationConfig{
			Enabled:   false,
			Mode:      "mcp",
			Status:    "not_configured",
			Fallback:  ".harness/artifacts/progress/decisions.md",
			HealthURL: "http://localhost:7437/health",
		},
		Stitch: IntegrationConfig{
			Enabled:  true,
			Mode:     "sdk",
			Status:   "not_configured",
			Fallback: "design-doc-only",
		},
		OpenDesign: IntegrationConfig{
			Enabled:  false,
			Mode:     "mcp",
			Status:   "not_configured",
			Fallback: "design-doc-only",
		},
		OpenPencil: IntegrationConfig{
			Enabled:  false,
			Mode:     "mcp",
			Status:   "not_configured",
			Fallback: "design-doc-only",
		},
	}
}

func (i *Integrations) ApplyPortableConfig(cfg *config.PortableConfig) {
	if i == nil || cfg == nil {
		return
	}
	cfg.Normalize()

	i.Engram.Mode = cfg.Integrations.Engram.Mode
	i.Engram.BinaryPath = cfg.Integrations.Engram.BinaryPath
	i.Engram.HealthURL = cfg.Integrations.Engram.HealthURL
	i.Engram.Fallback = cfg.Integrations.Engram.Fallback

	i.Stitch.Mode = cfg.Integrations.Stitch.Mode
	i.Stitch.Fallback = cfg.Integrations.Stitch.Fallback

	i.OpenDesign.Mode = cfg.Integrations.OpenDesign.Mode
	i.OpenDesign.MCPCommand = cfg.Integrations.OpenDesign.MCPCommand
	i.OpenDesign.MCPArgs = append([]string{}, cfg.Integrations.OpenDesign.MCPArgs...)
	i.OpenDesign.DataDir = cfg.Integrations.OpenDesign.DataDir
	i.OpenDesign.IPCPath = cfg.Integrations.OpenDesign.IPCPath
	i.OpenDesign.Fallback = cfg.Integrations.OpenDesign.Fallback

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
	i.Engram.LastDetectedAt = nowISO()
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
	i.OpenPencil.LastDetectedAt = nowISO()
}

func (i *Integrations) ApplyOpenDesignDetection(opendesign DetectionResult) {
	if i == nil {
		return
	}
	if i.Platform.OS == "" {
		i.Platform = opendesign.Platform
	}
	i.OpenDesign.Status = opendesign.Status
	i.OpenDesign.Reason = opendesign.Reason
	i.OpenDesign.Version = opendesign.Version
	i.OpenDesign.Fallback = opendesign.Fallback
	i.OpenDesign.LastDetectedAt = nowISO()
}

func (i *Integrations) ApplyStitchDetection(stitch DetectionResult) {
	if i == nil {
		return
	}
	if i.Platform.OS == "" {
		i.Platform = stitch.Platform
	}
	i.Stitch.Status = stitch.Status
	i.Stitch.Reason = stitch.Reason
	i.Stitch.Version = stitch.Version
	i.Stitch.Fallback = stitch.Fallback
	i.Stitch.LastDetectedAt = nowISO()
}

func (i *Integrations) Normalize() {
	if i == nil {
		return
	}
	defaults := DefaultIntegrations()
	if i.Engram.Mode == "" {
		i.Engram.Mode = defaults.Engram.Mode
	}
	if i.Engram.Status == "" {
		i.Engram.Status = defaults.Engram.Status
	}
	if i.Engram.Fallback == "" {
		i.Engram.Fallback = defaults.Engram.Fallback
	}
	if i.Engram.HealthURL == "" {
		i.Engram.HealthURL = defaults.Engram.HealthURL
	}
	stitchWasMissing := i.Stitch.Mode == "" && i.Stitch.Status == "" && i.Stitch.Fallback == ""
	if stitchWasMissing {
		i.Stitch.Enabled = defaults.Stitch.Enabled
	}
	if i.Stitch.Mode == "" {
		i.Stitch.Mode = defaults.Stitch.Mode
	}
	if i.Stitch.Status == "" {
		i.Stitch.Status = defaults.Stitch.Status
	}
	if i.Stitch.Fallback == "" {
		i.Stitch.Fallback = defaults.Stitch.Fallback
	}
	if i.OpenDesign.Mode == "" {
		i.OpenDesign.Mode = defaults.OpenDesign.Mode
	}
	if i.OpenDesign.Status == "" {
		i.OpenDesign.Status = defaults.OpenDesign.Status
	}
	if i.OpenDesign.Fallback == "" {
		i.OpenDesign.Fallback = defaults.OpenDesign.Fallback
	}
	if i.OpenPencil.Mode == "" {
		i.OpenPencil.Mode = defaults.OpenPencil.Mode
	}
	if i.OpenPencil.Status == "" {
		i.OpenPencil.Status = defaults.OpenPencil.Status
	}
	if i.OpenPencil.Fallback == "" {
		i.OpenPencil.Fallback = defaults.OpenPencil.Fallback
	}
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
	i.Normalize()
	return &i, nil
}

func (i *Integrations) Save() error {
	i.Normalize()
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

func (i *Integrations) EnableStitch() {
	i.Stitch.Enabled = true
	i.Stitch.Status = "enabled_via_cli"
}

func (i *Integrations) DisableStitch() {
	i.Stitch.Enabled = false
	i.Stitch.Status = "disabled_via_cli"
}

func (i *Integrations) IsStitchEnabled() bool {
	return i.Stitch.Enabled
}

func (i *Integrations) EnableOpenDesign() {
	i.OpenDesign.Enabled = true
	i.OpenDesign.Status = "enabled_via_cli"
}

func (i *Integrations) DisableOpenDesign() {
	i.OpenDesign.Enabled = false
	i.OpenDesign.Status = "disabled_via_cli"
}

func (i *Integrations) IsOpenDesignEnabled() bool {
	return i.OpenDesign.Enabled
}

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}
