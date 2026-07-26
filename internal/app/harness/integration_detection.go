package harness

import (
	config "shipwright/internal/config/application"
	integrations "shipwright/internal/integrations/application"
	platform "shipwright/internal/platform/application"
)

const (
	DetectionPathBinary    = integrations.DetectionPathBinary
	DetectionPathApp       = integrations.DetectionPathApp
	DetectionPathMCPServer = integrations.DetectionPathMCPServer

	DetectionNotInstalled         = integrations.DetectionNotInstalled
	DetectionInstalled            = integrations.DetectionInstalled
	DetectionAvailable            = integrations.DetectionAvailable
	DetectionConfiguredUnverified = integrations.DetectionConfiguredUnverified
	DetectionInstalledNoCanvas    = integrations.DetectionInstalledNoCanvas
	DetectionUnavailableFallback  = integrations.DetectionUnavailableFallback
)

type DetectionResult = integrations.DetectionResult

func DetectEngram(probe platform.SystemProbe) DetectionResult {
	return integrations.DetectEngram(probe)
}

func DetectEngramWithConfig(probe platform.SystemProbe, cfg *config.PortableConfig) DetectionResult {
	return integrations.DetectEngramWithConfig(probe, cfg)
}

func DetectOpenPencil(probe platform.SystemProbe) DetectionResult {
	return integrations.DetectOpenPencil(probe)
}

func DetectOpenPencilWithConfig(probe platform.SystemProbe, cfg *config.PortableConfig) DetectionResult {
	return integrations.DetectOpenPencilWithConfig(probe, cfg)
}

func DetectStitch(probe platform.SystemProbe) DetectionResult {
	return integrations.DetectStitch(probe)
}

func DetectStitchWithConfig(probe platform.SystemProbe, cfg *config.PortableConfig) DetectionResult {
	return integrations.DetectStitchWithConfig(probe, cfg)
}

func DetectOpenDesign(probe platform.SystemProbe) DetectionResult {
	return integrations.DetectOpenDesign(probe)
}

func DetectOpenDesignWithConfig(probe platform.SystemProbe, cfg *config.PortableConfig) DetectionResult {
	return integrations.DetectOpenDesignWithConfig(probe, cfg)
}
