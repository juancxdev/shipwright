package harness

import platform "shipwright/internal/platform/application"

type PlatformInfo = platform.PlatformInfo
type SystemProbe = platform.SystemProbe
type RealSystemProbe = platform.RealSystemProbe

func DetectPlatform(probe SystemProbe) PlatformInfo {
	return platform.DetectPlatform(probe)
}
