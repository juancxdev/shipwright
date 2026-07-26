package harness

import (
	config "shipwright/internal/config/application"
	platform "shipwright/internal/platform/application"
)

const PortableConfigFile = config.PortableConfigFile
const PortableConfigVersion = config.PortableConfigVersion

type PortableConfig = config.PortableConfig
type PortableHealthConfig = config.PortableHealthConfig
type PortableIntegrationsConfig = config.PortableIntegrationsConfig
type PortableExecutorsConfig = config.PortableExecutorsConfig
type PortableOpenCodeExecutorConfig = config.PortableOpenCodeExecutorConfig
type PortableEngramConfig = config.PortableEngramConfig
type PortableStitchConfig = config.PortableStitchConfig
type PortableOpenDesignConfig = config.PortableOpenDesignConfig
type PortableOpenPencilConfig = config.PortableOpenPencilConfig

func DefaultPortableConfig() *PortableConfig {
	return config.DefaultPortableConfig()
}

func LoadPortableConfigRaw() (*PortableConfig, error) {
	return config.LoadPortableConfigRaw()
}

func LoadPortableConfig() (*PortableConfig, error) {
	return config.LoadPortableConfig()
}

func LoadEffectivePortableConfig(probe platform.SystemProbe) (*PortableConfig, error) {
	return config.LoadEffectivePortableConfig(probe)
}

func DefaultOpenCodeExecutorConfig() PortableOpenCodeExecutorConfig {
	return config.DefaultOpenCodeExecutorConfig()
}
