package harness

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type DoctorFixResult struct {
	Applied bool     `json:"applied"`
	Actions []string `json:"actions"`
	Backup  string   `json:"backup,omitempty"`
}

func ApplyDoctorFixes(probe SystemProbe) (*DoctorFixResult, error) {
	if probe == nil {
		probe = RealSystemProbe{}
	}
	result := &DoctorFixResult{}

	_, statErr := probe.Stat(PortableConfigFile)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			cfg := DefaultPortableConfig()
			if err := cfg.Save(); err != nil {
				return result, err
			}
			result.Applied = true
			result.Actions = append(result.Actions, "created .harness/config.json with portable defaults")
			return result, nil
		}
		return result, statErr
	}

	cfg, err := LoadPortableConfigRaw()
	if err != nil {
		backup := corruptConfigBackupPath()
		if renameErr := os.Rename(PortableConfigFile, backup); renameErr != nil {
			return result, fmt.Errorf("cannot back up corrupt config: %w", renameErr)
		}
		if writeErr := DefaultPortableConfig().Save(); writeErr != nil {
			return result, fmt.Errorf("corrupt config backed up to %s but cannot write replacement: %w", backup, writeErr)
		}
		result.Applied = true
		result.Backup = backup
		result.Actions = append(result.Actions, "backed up corrupt .harness/config.json to "+backup)
		result.Actions = append(result.Actions, "recreated .harness/config.json with portable defaults")
		return result, nil
	}

	repairActions := RepairPortableConfig(cfg)
	if err := cfg.Save(); err != nil {
		return result, err
	}
	result.Applied = true
	if len(repairActions) == 0 {
		result.Actions = append(result.Actions, "normalized .harness/config.json defaults and missing fields")
	} else {
		result.Actions = append(result.Actions, repairActions...)
	}
	return result, nil
}

func RepairPortableConfig(cfg *PortableConfig) []string {
	if cfg == nil {
		return nil
	}
	var actions []string
	if strings.TrimSpace(cfg.Version) != PortableConfigVersion {
		cfg.Version = PortableConfigVersion
		actions = append(actions, "set config version to "+PortableConfigVersion)
	}
	if strings.TrimSpace(cfg.ArtifactRoot) == "" {
		cfg.ArtifactRoot = "."
		actions = append(actions, "set artifact_root to .")
	}
	if cfg.Health.TimeoutMillis <= 0 {
		cfg.Health.TimeoutMillis = DefaultHealthTimeoutMillis
		actions = append(actions, fmt.Sprintf("set health.timeout_millis to %d", DefaultHealthTimeoutMillis))
	}
	if strings.TrimSpace(cfg.Integrations.Engram.Mode) == "" || !isSupportedConfigMode(cfg.Integrations.Engram.Mode) {
		cfg.Integrations.Engram.Mode = ConfigModeMCP
		actions = append(actions, "set integrations.engram.mode to mcp")
	}
	if strings.TrimSpace(cfg.Integrations.OpenPencil.Mode) == "" || !isSupportedConfigMode(cfg.Integrations.OpenPencil.Mode) {
		cfg.Integrations.OpenPencil.Mode = ConfigModeMCP
		actions = append(actions, "set integrations.openpencil.mode to mcp")
	}
	if strings.TrimSpace(cfg.Integrations.Engram.Fallback) == "" {
		cfg.Integrations.Engram.Fallback = DecisionsFile
		actions = append(actions, "set integrations.engram.fallback")
	}
	if strings.TrimSpace(cfg.Integrations.OpenPencil.Fallback) == "" {
		cfg.Integrations.OpenPencil.Fallback = "design-doc-only"
		actions = append(actions, "set integrations.openpencil.fallback")
	}
	if validateHTTPURL("integrations.engram.health_url", cfg.Integrations.Engram.HealthURL) != nil {
		cfg.Integrations.Engram.HealthURL = "http://localhost:7437/health"
		actions = append(actions, "reset integrations.engram.health_url to default")
	}
	for osName, override := range cfg.PlatformOverrides {
		changed := false
		if strings.TrimSpace(override.Engram.Mode) != "" && !isSupportedConfigMode(override.Engram.Mode) {
			override.Engram.Mode = ConfigModeMCP
			changed = true
			actions = append(actions, "set platform_overrides."+osName+".engram.mode to mcp")
		}
		if strings.TrimSpace(override.OpenPencil.Mode) != "" && !isSupportedConfigMode(override.OpenPencil.Mode) {
			override.OpenPencil.Mode = ConfigModeMCP
			changed = true
			actions = append(actions, "set platform_overrides."+osName+".openpencil.mode to mcp")
		}
		if strings.TrimSpace(override.Engram.HealthURL) != "" && validateHTTPURL("platform_overrides."+osName+".engram.health_url", override.Engram.HealthURL) != nil {
			override.Engram.HealthURL = ""
			changed = true
			actions = append(actions, "removed invalid platform_overrides."+osName+".engram.health_url")
		}
		if changed {
			cfg.PlatformOverrides[osName] = override
		}
	}
	cfg.Normalize()
	return actions
}

func isSupportedConfigMode(mode string) bool {
	mode = strings.TrimSpace(mode)
	return mode == ConfigModeMCP || mode == ConfigModeDisabled
}

func corruptConfigBackupPath() string {
	replacer := strings.NewReplacer(":", "-", ".", "-", "T", "-", "Z", "")
	stamp := replacer.Replace(NowISO())
	return PortableConfigFile + ".corrupt." + stamp + ".bak"
}
