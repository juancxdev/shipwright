package harness

import (
	config "shipwright/internal/config/application"
	integrations "shipwright/internal/integrations/application"
	"time"
)

const DefaultHealthTimeoutMillis = integrations.DefaultHealthTimeoutMillis

const (
	HealthStatusSkipped   = integrations.HealthStatusSkipped
	HealthStatusHealthy   = integrations.HealthStatusHealthy
	HealthStatusUnhealthy = integrations.HealthStatusUnhealthy
)

type HealthResult = integrations.HealthResult
type HealthProbe = integrations.HealthProbe
type RealHealthProbe = integrations.RealHealthProbe

func HealthTimeout(cfg *config.PortableConfig) time.Duration {
	return integrations.HealthTimeout(cfg)
}

func CheckEngramHealth(probe HealthProbe, cfg *config.PortableConfig, detected DetectionResult) HealthResult {
	return integrations.CheckEngramHealth(probe, cfg, detected)
}

func CheckOpenPencilHealth(probe HealthProbe, cfg *config.PortableConfig, detected DetectionResult) HealthResult {
	return integrations.CheckOpenPencilHealth(probe, cfg, detected)
}
