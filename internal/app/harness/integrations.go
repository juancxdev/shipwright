package harness

import (
	integrations "shipwright/internal/integrations/application"
)

const IntegrationsFile = integrations.IntegrationsFile

type IntegrationConfig = integrations.IntegrationConfig
type Integrations = integrations.Integrations

func DefaultIntegrations() *Integrations {
	return integrations.DefaultIntegrations()
}

func LoadIntegrations() (*Integrations, error) {
	return integrations.LoadIntegrations()
}
