package harness

import (
	"errors"
	"fmt"
	"os"
)

const (
	DoctorSeverityOK      = "OK"
	DoctorSeverityInfo    = "INFO"
	DoctorSeverityWarning = "WARNING"
	DoctorSeverityError   = "ERROR"
)

type DoctorReport struct {
	Platform         PlatformInfo            `json:"platform"`
	ConfigFile       string                  `json:"config_file"`
	ConfigExists     bool                    `json:"config_exists"`
	ConfigLoaded     bool                    `json:"config_loaded"`
	Engram           DetectionResult         `json:"engram"`
	OpenPencil       DetectionResult         `json:"openpencil"`
	EngramHealth     HealthResult            `json:"engram_health"`
	OpenPencilHealth HealthResult            `json:"openpencil_health"`
	ConfigIssues     []ConfigValidationIssue `json:"config_issues,omitempty"`
	Checks           []DoctorCheck           `json:"checks"`
	Actions          []string                `json:"actions"`
	Summary          DoctorSummary           `json:"summary"`
}

type DoctorCheck struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Severity string `json:"severity"`
	Status   string `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Action   string `json:"action,omitempty"`
}

type DoctorSummary struct {
	OK       int  `json:"ok"`
	Info     int  `json:"info"`
	Warnings int  `json:"warnings"`
	Errors   int  `json:"errors"`
	Healthy  bool `json:"healthy"`
}

func RunDoctor(probe SystemProbe) (*DoctorReport, error) {
	return RunDoctorWithHealth(probe, RealHealthProbe{})
}

func RunDoctorWithHealth(probe SystemProbe, healthProbe HealthProbe) (*DoctorReport, error) {
	if probe == nil {
		probe = RealSystemProbe{}
	}
	if healthProbe == nil {
		healthProbe = RealHealthProbe{}
	}

	platform := DetectPlatform(probe)
	report := &DoctorReport{
		Platform:   platform,
		ConfigFile: PortableConfigFile,
	}

	cfg, cfgExists, cfgErr := loadDoctorConfig(probe)
	report.ConfigExists = cfgExists
	report.ConfigLoaded = cfgErr == nil

	if cfgErr != nil {
		report.addCheck(DoctorCheck{
			ID:       "config.load",
			Title:    "Portable config loads",
			Severity: DoctorSeverityError,
			Status:   "failed",
			Detail:   cfgErr.Error(),
			Action:   "Fix .harness/config.json or recreate it with 'shipwright config init'.",
		})
		cfg = DefaultPortableConfig()
	} else if cfgExists {
		report.addCheck(DoctorCheck{
			ID:       "config.load",
			Title:    "Portable config loads",
			Severity: DoctorSeverityOK,
			Status:   "ok",
			Detail:   ".harness/config.json loaded successfully",
		})
		report.evaluateConfigValidation()
	} else {
		report.addCheck(DoctorCheck{
			ID:       "config.exists",
			Title:    "Portable config exists",
			Severity: DoctorSeverityWarning,
			Status:   "missing",
			Detail:   ".harness/config.json not found; defaults are being used",
			Action:   "Run 'shipwright config init' to create a portable project config.",
		})
	}

	report.addCheck(DoctorCheck{
		ID:       "platform.detect",
		Title:    "Platform detected",
		Severity: DoctorSeverityOK,
		Status:   "ok",
		Detail:   fmt.Sprintf("%s/%s (ci=%s)", platform.OS, platform.Arch, formatYesNo(platform.IsCI)),
	})

	report.Engram = DetectEngramWithConfig(probe, cfg)
	report.OpenPencil = DetectOpenPencilWithConfig(probe, cfg)
	report.EngramHealth = CheckEngramHealth(healthProbe, cfg, report.Engram)
	report.OpenPencilHealth = CheckOpenPencilHealth(healthProbe, cfg, report.OpenPencil)
	report.evaluateEngram(cfg)
	report.evaluateOpenPencil(cfg)
	report.evaluateHealth()
	report.Summary = summarizeDoctorChecks(report.Checks)
	report.Summary.Healthy = report.Summary.Errors == 0
	return report, nil
}

func (r *DoctorReport) evaluateConfigValidation() {
	raw, err := LoadPortableConfigRaw()
	if err != nil {
		r.addCheck(DoctorCheck{
			ID:       "config.validation",
			Title:    "Portable config validates",
			Severity: DoctorSeverityError,
			Status:   "failed",
			Detail:   err.Error(),
			Action:   "Run 'shipwright doctor --fix' to back up and recreate config.",
		})
		return
	}
	issues := ValidatePortableConfig(raw)
	r.ConfigIssues = issues
	if len(issues) == 0 {
		r.addCheck(DoctorCheck{
			ID:       "config.validation",
			Title:    "Portable config validates",
			Severity: DoctorSeverityOK,
			Status:   "ok",
			Detail:   "semantic config validation passed",
		})
		return
	}
	for _, issue := range issues {
		severity := DoctorSeverityWarning
		if issue.Severity == ConfigIssueSeverityError {
			severity = DoctorSeverityError
		}
		r.addCheck(DoctorCheck{
			ID:       issue.ID,
			Title:    "Portable config issue",
			Severity: severity,
			Status:   "invalid",
			Detail:   fmt.Sprintf("%s: %s", issue.Path, issue.Message),
			Action:   issue.Action,
		})
	}
}

func loadDoctorConfig(probe SystemProbe) (*PortableConfig, bool, error) {
	_, statErr := probe.Stat(PortableConfigFile)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			cfg := DefaultPortableConfig()
			cfg.ApplyEnv(probe)
			cfg.Normalize()
			return cfg, false, nil
		}
		return nil, false, statErr
	}

	cfg, err := LoadEffectivePortableConfig(probe)
	if err != nil {
		return nil, true, err
	}
	return cfg, true, nil
}

func (r *DoctorReport) evaluateEngram(cfg *PortableConfig) {
	if r.Engram.Available {
		r.addCheck(DoctorCheck{
			ID:       "engram.available",
			Title:    "Engram available",
			Severity: DoctorSeverityOK,
			Status:   "available",
			Detail:   pathDetail(r.Engram),
		})
		return
	}

	severity := DoctorSeverityInfo
	if r.Engram.Configured {
		severity = DoctorSeverityWarning
	}
	action := "Shipwright will use memory fallback at " + r.Engram.Fallback + "."
	if r.Engram.Configured {
		action = "Fix the configured Engram binary path or unset ENGRAM_BINARY / config.integrations.engram.binary_path."
	}
	r.addCheck(DoctorCheck{
		ID:       "engram.available",
		Title:    "Engram available",
		Severity: severity,
		Status:   r.Engram.Status,
		Detail:   r.Engram.Reason,
		Action:   action,
	})
}

func (r *DoctorReport) evaluateOpenPencil(cfg *PortableConfig) {
	if r.OpenPencil.Available && r.OpenPencil.Active {
		r.addCheck(DoctorCheck{
			ID:       "openpencil.available",
			Title:    "OpenPencil available",
			Severity: DoctorSeverityOK,
			Status:   "available",
			Detail:   pathDetail(r.OpenPencil),
		})
		return
	}

	if r.OpenPencil.Installed && !r.OpenPencil.Active {
		r.addCheck(DoctorCheck{
			ID:       "openpencil.canvas",
			Title:    "OpenPencil canvas active",
			Severity: DoctorSeverityWarning,
			Status:   r.OpenPencil.Status,
			Detail:   r.OpenPencil.Reason,
			Action:   "Open OpenPencil and activate a canvas, or continue with fallback: " + r.OpenPencil.Fallback + ".",
		})
		return
	}

	severity := DoctorSeverityInfo
	if r.OpenPencil.Configured {
		severity = DoctorSeverityWarning
	}
	action := "Shipwright will use design fallback: " + r.OpenPencil.Fallback + "."
	if r.OpenPencil.Configured {
		action = "Fix OPENPENCIL_APP_PATH / OPENPENCIL_MCP_SERVER or the matching .harness/config.json value."
	}
	r.addCheck(DoctorCheck{
		ID:       "openpencil.available",
		Title:    "OpenPencil available",
		Severity: severity,
		Status:   r.OpenPencil.Status,
		Detail:   r.OpenPencil.Reason,
		Action:   action,
	})
}

func (r *DoctorReport) evaluateHealth() {
	r.evaluateHealthResult("engram.health", "Engram health check", r.EngramHealth)
	r.evaluateHealthResult("openpencil.health", "OpenPencil MCP health check", r.OpenPencilHealth)
}

func (r *DoctorReport) evaluateHealthResult(id, title string, result HealthResult) {
	if !result.Checked || result.Status == HealthStatusSkipped {
		r.addCheck(DoctorCheck{
			ID:       id,
			Title:    title,
			Severity: DoctorSeverityInfo,
			Status:   HealthStatusSkipped,
			Detail:   result.Detail,
		})
		return
	}
	if result.Healthy {
		r.addCheck(DoctorCheck{
			ID:       id,
			Title:    title,
			Severity: DoctorSeverityOK,
			Status:   HealthStatusHealthy,
			Detail:   healthDetail(result),
		})
		return
	}
	r.addCheck(DoctorCheck{
		ID:       id,
		Title:    title,
		Severity: DoctorSeverityWarning,
		Status:   HealthStatusUnhealthy,
		Detail:   healthDetail(result),
		Action:   result.Suggestion,
	})
}

func healthDetail(result HealthResult) string {
	detail := result.Detail
	if result.Endpoint != "" {
		detail = fmt.Sprintf("%s — %s", result.Endpoint, detail)
	}
	if result.LatencyMS > 0 {
		detail = fmt.Sprintf("%s (%dms)", detail, result.LatencyMS)
	}
	return detail
}

func (r *DoctorReport) addCheck(check DoctorCheck) {
	r.Checks = append(r.Checks, check)
	if check.Action != "" {
		r.Actions = append(r.Actions, check.Action)
	}
}

func summarizeDoctorChecks(checks []DoctorCheck) DoctorSummary {
	var summary DoctorSummary
	for _, check := range checks {
		switch check.Severity {
		case DoctorSeverityOK:
			summary.OK++
		case DoctorSeverityWarning:
			summary.Warnings++
		case DoctorSeverityError:
			summary.Errors++
		default:
			summary.Info++
		}
	}
	return summary
}

func pathDetail(result DetectionResult) string {
	if result.Path == "" {
		return result.Reason
	}
	if result.PathKind == "" {
		return result.Path
	}
	return fmt.Sprintf("%s: %s", result.PathKind, result.Path)
}
