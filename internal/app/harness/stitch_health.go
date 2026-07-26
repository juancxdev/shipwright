package harness

import (
	"context"
	"net/http"
	integrations "shipwright/internal/integrations/application"
)

const DefaultStitchMCPEndpoint = integrations.DefaultStitchMCPEndpoint

type StitchConnectionResult = integrations.StitchConnectionResult

func ValidateStitchConnection(ctx context.Context, client *http.Client, endpoint string, apiKey string) StitchConnectionResult {
	return integrations.ValidateStitchConnection(ctx, client, endpoint, apiKey)
}
