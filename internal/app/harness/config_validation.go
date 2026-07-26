package harness

import config "shipwright/internal/config/application"

const (
	ConfigIssueSeverityError   = config.ConfigIssueSeverityError
	ConfigIssueSeverityWarning = config.ConfigIssueSeverityWarning
)

const (
	ConfigModeMCP      = config.ConfigModeMCP
	ConfigModeSDK      = config.ConfigModeSDK
	ConfigModeDisabled = config.ConfigModeDisabled
)

type ConfigValidationIssue = config.ConfigValidationIssue

func ValidatePortableConfig(cfg *PortableConfig) []ConfigValidationIssue {
	return config.ValidatePortableConfig(cfg)
}

func CountConfigIssueSeverities(issues []ConfigValidationIssue) (errorsCount int, warningsCount int) {
	return config.CountConfigIssueSeverities(issues)
}

func validateHTTPURL(path, raw string) *ConfigValidationIssue {
	return config.ValidateHTTPURL(path, raw)
}
