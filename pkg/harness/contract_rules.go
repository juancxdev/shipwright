package harness

import (
	"fmt"
	"os"
	"strings"
)

type MockCompliance struct {
	HasMockMode      bool
	HasHTTPMode      bool
	MocksPreserved   bool
	HTTPSeparate     bool
	EndpointCoverage []MockEndpointCoverage
	Issues           []string
	Warnings         []string
}

type MockEndpointCoverage struct {
	Endpoint  string
	HasMock   bool
	HasHTTP   bool
	HasToggle bool
}

type BackendCompliance struct {
	ContractExists        bool
	EndpointsMatch        []EndpointMatch
	SchemasMatch          []SchemaMatch
	ErrorFormatConsistent bool
	Issues                []string
	Warnings              []string
}

type EndpointMatch struct {
	Endpoint   string
	InContract bool
	InProgress bool
	Status     string
}

type SchemaMatch struct {
	Schema     string
	InContract bool
	InProgress bool
	Status     string
}

func CheckMockCompliance(spec *ContractSpec) *MockCompliance {
	result := &MockCompliance{}

	feProgress, err := os.ReadFile("progress/frontend.md")
	if err != nil {
		result.Issues = append(result.Issues, "progress/frontend.md not found — cannot verify mock compliance")
		return result
	}

	feContent := strings.ToLower(string(feProgress))

	result.HasMockMode = strings.Contains(feContent, "mock") || strings.Contains(feContent, "mock mode")
	result.HasHTTPMode = strings.Contains(feContent, "http") || strings.Contains(feContent, "real api") || strings.Contains(feContent, "http mode")
	result.MocksPreserved = !strings.Contains(feContent, "mock deleted") && !strings.Contains(feContent, "mock removed")
	result.HTTPSeparate = result.HasMockMode && result.HasHTTPMode

	for _, ep := range spec.Endpoints {
		coverage := MockEndpointCoverage{
			Endpoint: fmt.Sprintf("%s %s", ep.Method, ep.Path),
		}

		epLower := strings.ToLower(ep.Path)
		coverage.HasMock = strings.Contains(feContent, "mock") && (strings.Contains(feContent, epLower) ||
			strings.Contains(feContent, extractResourceName(ep.Path)))

		coverage.HasHTTP = strings.Contains(feContent, "http") && (strings.Contains(feContent, epLower) ||
			strings.Contains(feContent, extractResourceName(ep.Path)))

		coverage.HasToggle = strings.Contains(feContent, "toggle") || strings.Contains(feContent, "mode switch")

		result.EndpointCoverage = append(result.EndpointCoverage, coverage)
	}

	if !result.HasMockMode {
		result.Issues = append(result.Issues, "Frontend progress does not mention mock mode — mocks are MANDATORY")
	}
	if !result.HasHTTPMode {
		result.Issues = append(result.Issues, "Frontend progress does not mention HTTP/real API mode")
	}
	if !result.MocksPreserved {
		result.Issues = append(result.Issues, "Frontend progress indicates mocks were deleted — mocks MUST be preserved")
	}
	if !result.HTTPSeparate {
		result.Warnings = append(result.Warnings, "Mock and HTTP modes may not be properly separated")
	}

	return result
}

func CheckBackendCompliance(spec *ContractSpec) *BackendCompliance {
	result := &BackendCompliance{}

	if !ContractExists() {
		result.Issues = append(result.Issues, "contracts/openapi.yaml does not exist")
		return result
	}

	beProgress, err := os.ReadFile("progress/backend.md")
	if err != nil {
		result.Issues = append(result.Issues, "progress/backend.md not found — cannot verify backend compliance")
		return result
	}

	beContent := strings.ToLower(string(beProgress))
	result.ContractExists = true
	result.ErrorFormatConsistent = strings.Contains(beContent, "error") && (strings.Contains(beContent, "consistent") ||
		strings.Contains(beContent, "format") ||
		strings.Contains(beContent, "400") ||
		strings.Contains(beContent, "401") ||
		strings.Contains(beContent, "500"))

	for _, ep := range spec.Endpoints {
		match := EndpointMatch{
			Endpoint:   fmt.Sprintf("%s %s", ep.Method, ep.Path),
			InContract: true,
		}

		epLower := strings.ToLower(ep.Path)
		match.InProgress = strings.Contains(beContent, epLower) ||
			strings.Contains(beContent, strings.ToLower(ep.Method)) && strings.Contains(beContent, epLower)

		if match.InProgress {
			match.Status = "covered"
		} else {
			match.Status = "missing"
			result.Warnings = append(result.Warnings, fmt.Sprintf("Backend progress does not mention: %s %s", ep.Method, ep.Path))
		}

		result.EndpointsMatch = append(result.EndpointsMatch, match)
	}

	for _, s := range spec.Schemas {
		match := SchemaMatch{
			Schema:     s.Name,
			InContract: true,
		}

		match.InProgress = strings.Contains(beContent, strings.ToLower(s.Name))
		if match.InProgress {
			match.Status = "covered"
		} else {
			match.Status = "missing"
			result.Warnings = append(result.Warnings, fmt.Sprintf("Backend progress does not mention schema: %s", s.Name))
		}

		result.SchemasMatch = append(result.SchemasMatch, match)
	}

	if !result.ErrorFormatConsistent {
		result.Issues = append(result.Issues, "Backend progress does not document consistent error response format")
	}

	return result
}

func FormatMockCompliance(mc *MockCompliance) string {
	var sb strings.Builder

	sb.WriteString("Mock Compliance Check\n")
	sb.WriteString("=====================\n\n")

	sb.WriteString(fmt.Sprintf("Mock mode:       %s\n", formatYesNo(mc.HasMockMode)))
	sb.WriteString(fmt.Sprintf("HTTP mode:       %s\n", formatYesNo(mc.HasHTTPMode)))
	sb.WriteString(fmt.Sprintf("Mocks preserved: %s\n", formatYesNo(mc.MocksPreserved)))
	sb.WriteString(fmt.Sprintf("HTTP separate:   %s\n\n", formatYesNo(mc.HTTPSeparate)))

	if len(mc.EndpointCoverage) > 0 {
		sb.WriteString("Endpoint coverage:\n")
		sb.WriteString(fmt.Sprintf("  %-35s  %-8s  %-8s  %-8s\n", "ENDPOINT", "MOCK", "HTTP", "TOGGLE"))
		for _, c := range mc.EndpointCoverage {
			sb.WriteString(fmt.Sprintf("  %-35s  %-8s  %-8s  %-8s\n",
				c.Endpoint, formatYesNo(c.HasMock), formatYesNo(c.HasHTTP), formatYesNo(c.HasToggle)))
		}
		sb.WriteString("\n")
	}

	if len(mc.Issues) > 0 {
		sb.WriteString("Issues:\n")
		for _, i := range mc.Issues {
			sb.WriteString(fmt.Sprintf("  ✗ %s\n", i))
		}
	}

	if len(mc.Warnings) > 0 {
		sb.WriteString("\nWarnings:\n")
		for _, w := range mc.Warnings {
			sb.WriteString(fmt.Sprintf("  ⚠ %s\n", w))
		}
	}

	if len(mc.Issues) == 0 {
		sb.WriteString("\n✓ Mock compliance: PASS\n")
	} else {
		sb.WriteString("\n✗ Mock compliance: FAIL\n")
	}

	return sb.String()
}

func FormatBackendCompliance(bc *BackendCompliance) string {
	var sb strings.Builder

	sb.WriteString("Backend Contract Compliance\n")
	sb.WriteString("============================\n\n")

	sb.WriteString(fmt.Sprintf("Contract exists:     %s\n", formatYesNo(bc.ContractExists)))
	sb.WriteString(fmt.Sprintf("Error format:        %s\n\n", formatYesNo(bc.ErrorFormatConsistent)))

	if len(bc.EndpointsMatch) > 0 {
		sb.WriteString("Endpoint coverage:\n")
		sb.WriteString(fmt.Sprintf("  %-35s  %-10s  %-10s  %s\n", "ENDPOINT", "CONTRACT", "PROGRESS", "STATUS"))
		for _, m := range bc.EndpointsMatch {
			sb.WriteString(fmt.Sprintf("  %-35s  %-10s  %-10s  %s\n",
				m.Endpoint, formatYesNo(m.InContract), formatYesNo(m.InProgress), m.Status))
		}
		sb.WriteString("\n")
	}

	if len(bc.SchemasMatch) > 0 {
		sb.WriteString("Schema coverage:\n")
		sb.WriteString(fmt.Sprintf("  %-25s  %-10s  %-10s  %s\n", "SCHEMA", "CONTRACT", "PROGRESS", "STATUS"))
		for _, m := range bc.SchemasMatch {
			sb.WriteString(fmt.Sprintf("  %-25s  %-10s  %-10s  %s\n",
				m.Schema, formatYesNo(m.InContract), formatYesNo(m.InProgress), m.Status))
		}
		sb.WriteString("\n")
	}

	if len(bc.Issues) > 0 {
		sb.WriteString("Issues:\n")
		for _, i := range bc.Issues {
			sb.WriteString(fmt.Sprintf("  ✗ %s\n", i))
		}
	}

	if len(bc.Warnings) > 0 {
		sb.WriteString("\nWarnings:\n")
		for _, w := range bc.Warnings {
			sb.WriteString(fmt.Sprintf("  ⚠ %s\n", w))
		}
	}

	if len(bc.Issues) == 0 {
		sb.WriteString("\n✓ Backend compliance: PASS\n")
	} else {
		sb.WriteString("\n✗ Backend compliance: FAIL\n")
	}

	return sb.String()
}
