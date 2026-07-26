package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultStitchMCPEndpoint = "https://stitch.googleapis.com/mcp"

type StitchConnectionResult struct {
	Checked bool   `json:"checked"`
	Healthy bool   `json:"healthy"`
	Target  string `json:"target,omitempty"`
	Status  string `json:"status"`
	Detail  string `json:"detail,omitempty"`
}

type stitchInitializeRequest struct {
	JSONRPC string                        `json:"jsonrpc"`
	ID      int                           `json:"id"`
	Method  string                        `json:"method"`
	Params  stitchInitializeRequestParams `json:"params"`
}

type stitchInitializeRequestParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]any         `json:"capabilities"`
	ClientInfo      stitchInitializeClient `json:"clientInfo"`
}

type stitchInitializeClient struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func ValidateStitchConnection(ctx context.Context, client *http.Client, endpoint string, apiKey string) StitchConnectionResult {
	endpoint = strings.TrimSpace(endpoint)
	apiKey = strings.TrimSpace(apiKey)
	if endpoint == "" {
		endpoint = DefaultStitchMCPEndpoint
	}
	if apiKey == "" {
		return StitchConnectionResult{
			Checked: false,
			Healthy: false,
			Target:  endpoint,
			Status:  "skipped",
			Detail:  "STITCH_API_KEY is empty",
		}
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	payload := stitchInitializeRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: stitchInitializeRequestParams{
			ProtocolVersion: "2025-03-26",
			Capabilities:    map[string]any{},
			ClientInfo: stitchInitializeClient{
				Name:    "shipwright",
				Version: "init-wizard",
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return StitchConnectionResult{Checked: true, Healthy: false, Target: endpoint, Status: "error", Detail: err.Error()}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return StitchConnectionResult{Checked: true, Healthy: false, Target: endpoint, Status: "error", Detail: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("X-Goog-Api-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return StitchConnectionResult{Checked: true, Healthy: false, Target: endpoint, Status: "error", Detail: err.Error()}
	}
	defer resp.Body.Close()
	responseBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	detail := strings.TrimSpace(string(responseBytes))
	if len(detail) > 180 {
		detail = detail[:180] + "…"
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return StitchConnectionResult{Checked: true, Healthy: true, Target: endpoint, Status: "healthy", Detail: fmt.Sprintf("%s %s", resp.Status, detail)}
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return StitchConnectionResult{Checked: true, Healthy: false, Target: endpoint, Status: "auth_failed", Detail: resp.Status}
	}
	return StitchConnectionResult{Checked: true, Healthy: false, Target: endpoint, Status: "unhealthy", Detail: fmt.Sprintf("%s %s", resp.Status, detail)}
}
