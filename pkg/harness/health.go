package harness

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const DefaultHealthTimeoutMillis = 1500

const (
	HealthStatusSkipped   = "skipped"
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
)

type HealthResult struct {
	Name       string `json:"name"`
	Checked    bool   `json:"checked"`
	Healthy    bool   `json:"healthy"`
	Status     string `json:"status"`
	Endpoint   string `json:"endpoint,omitempty"`
	LatencyMS  int64  `json:"latency_ms,omitempty"`
	Detail     string `json:"detail,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

type HealthProbe interface {
	CheckHTTP(url string, timeout time.Duration) HealthResult
	CheckCommand(name string, args []string, timeout time.Duration) HealthResult
}

type RealHealthProbe struct{}

func (RealHealthProbe) CheckHTTP(url string, timeout time.Duration) HealthResult {
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return HealthResult{Name: "http", Checked: true, Status: HealthStatusUnhealthy, Endpoint: url, Detail: err.Error()}
	}

	resp, err := http.DefaultClient.Do(req)
	latency := time.Since(started).Milliseconds()
	if err != nil {
		return HealthResult{Name: "http", Checked: true, Status: HealthStatusUnhealthy, Endpoint: url, LatencyMS: latency, Detail: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return HealthResult{Name: "http", Checked: true, Healthy: true, Status: HealthStatusHealthy, Endpoint: url, LatencyMS: latency, Detail: resp.Status}
	}
	return HealthResult{Name: "http", Checked: true, Status: HealthStatusUnhealthy, Endpoint: url, LatencyMS: latency, Detail: resp.Status}
}

func (RealHealthProbe) CheckCommand(name string, args []string, timeout time.Duration) HealthResult {
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	latency := time.Since(started).Milliseconds()
	detail := strings.TrimSpace(string(output))
	if ctx.Err() == context.DeadlineExceeded {
		return HealthResult{Name: name, Checked: true, Status: HealthStatusUnhealthy, LatencyMS: latency, Detail: "command timed out"}
	}
	if err != nil {
		if detail == "" {
			detail = err.Error()
		} else {
			detail = fmt.Sprintf("%s: %s", err, detail)
		}
		return HealthResult{Name: name, Checked: true, Status: HealthStatusUnhealthy, LatencyMS: latency, Detail: detail}
	}
	if detail == "" {
		detail = "command exited successfully"
	}
	return HealthResult{Name: name, Checked: true, Healthy: true, Status: HealthStatusHealthy, LatencyMS: latency, Detail: detail}
}

func HealthTimeout(cfg *PortableConfig) time.Duration {
	millis := DefaultHealthTimeoutMillis
	if cfg != nil && cfg.Health.TimeoutMillis > 0 {
		millis = cfg.Health.TimeoutMillis
	}
	return time.Duration(millis) * time.Millisecond
}

func CheckEngramHealth(probe HealthProbe, cfg *PortableConfig, detected DetectionResult) HealthResult {
	if probe == nil {
		probe = RealHealthProbe{}
	}
	if cfg == nil || cfg.Integrations.Engram.HealthURL == "" {
		return HealthResult{Name: "engram", Status: HealthStatusSkipped, Detail: "health URL not configured"}
	}
	if !detected.Available {
		return HealthResult{Name: "engram", Status: HealthStatusSkipped, Detail: "engram binary is not available"}
	}
	result := probe.CheckHTTP(cfg.Integrations.Engram.HealthURL, HealthTimeout(cfg))
	result.Name = "engram"
	if !result.Healthy {
		result.Suggestion = "Start Engram health service or update ENGRAM_HEALTH_URL / config.integrations.engram.health_url."
	}
	return result
}

func CheckOpenPencilHealth(probe HealthProbe, cfg *PortableConfig, detected DetectionResult) HealthResult {
	if probe == nil {
		probe = RealHealthProbe{}
	}
	if !detected.Installed || detected.Path == "" {
		return HealthResult{Name: "openpencil", Status: HealthStatusSkipped, Detail: "OpenPencil MCP server path/command is not available"}
	}
	if detected.PathKind == DetectionPathBinary {
		return HealthResult{Name: "openpencil", Status: HealthStatusSkipped, Endpoint: detected.Path, Detail: "OpenPencil MCP command found; stdio server health is verified by the MCP client with the OpenPencil app open"}
	}
	if detected.PathKind != DetectionPathMCPServer {
		return HealthResult{Name: "openpencil", Status: HealthStatusSkipped, Detail: "OpenPencil MCP server path is not available"}
	}
	result := probe.CheckCommand("node", []string{"--check", detected.Path}, HealthTimeout(cfg))
	result.Name = "openpencil"
	result.Endpoint = detected.Path
	if !result.Healthy {
		result.Suggestion = "Verify Node.js is installed and OPENPENCIL_MCP_SERVER points to a readable JavaScript MCP server, or install @open-pencil/mcp and use OPENPENCIL_MCP_COMMAND=openpencil-mcp."
	}
	return result
}
