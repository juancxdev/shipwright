package harness

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

const ContractFile = "contracts/openapi.yaml"

type ContractEndpoint struct {
	Path       string
	Method     string
	Summary    string
	HasAuth    bool
	HasRequest bool
	Responses  []string
}

type ContractSchema struct {
	Name       string
	Properties []string
}

type ContractSpec struct {
	Title         string
	Version       string
	HasAuth       bool
	Endpoints     []ContractEndpoint
	Schemas       []ContractSchema
	EndpointCount int
	SchemaCount   int
}

type ContractParseResult struct {
	Spec     *ContractSpec
	Warnings []string
	Errors   []string
	IsValid  bool
}

func ParseContract(path string) *ContractParseResult {
	result := &ContractParseResult{}

	data, err := os.ReadFile(path)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("cannot read %s: %s", path, err))
		return result
	}

	content := string(data)
	spec := &ContractSpec{}

	spec.Title = extractYAMLValue(content, "title")
	spec.Version = extractYAMLValue(content, "version")

	if spec.Title == "" {
		result.Warnings = append(result.Warnings, "missing info.title")
	}
	if spec.Version == "" {
		result.Warnings = append(result.Warnings, "missing info.version")
	}

	authSchemePattern := regexp.MustCompile(`securitySchemes:\s*\n\s+(\w+):\s*\n\s+type:\s*(\w+)`)
	if authSchemePattern.MatchString(content) {
		spec.HasAuth = true
	}

	spec.Endpoints = extractEndpoints(content)
	spec.EndpointCount = len(spec.Endpoints)

	spec.Schemas = extractSchemas(content)
	spec.SchemaCount = len(spec.Schemas)

	for _, ep := range spec.Endpoints {
		if ep.Summary == "" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s %s: missing summary", ep.Method, ep.Path))
		}
		hasErrorResp := false
		for _, r := range ep.Responses {
			if strings.HasPrefix(r, "4") || strings.HasPrefix(r, "5") {
				hasErrorResp = true
				break
			}
		}
		if !hasErrorResp {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s %s: no error responses (4xx/5xx)", ep.Method, ep.Path))
		}
		if spec.HasAuth && !ep.HasAuth && !isPublicEndpoint(ep.Path) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s %s: contract has auth scheme but endpoint has no security", ep.Method, ep.Path))
		}
	}

	if spec.EndpointCount == 0 {
		hasPathsSection := strings.Contains(content, "paths:")
		if hasPathsSection {
			result.Warnings = append(result.Warnings, "paths: section exists but no endpoints were parsed (check indentation)")
		} else {
			result.Errors = append(result.Errors, "no paths: section found — contract defines no endpoints")
		}
	}

	if len(result.Errors) == 0 {
		result.IsValid = true
	}

	result.Spec = spec
	return result
}

func extractYAMLValue(content, key string) string {
	pattern := regexp.MustCompile(`(?m)^\s+` + key + `:\s*(.+)$`)
	match := pattern.FindStringSubmatch(content)
	if len(match) >= 2 {
		val := strings.TrimSpace(match[1])
		val = strings.Trim(val, "\"'")
		return val
	}
	return ""
}

func extractEndpoints(content string) []ContractEndpoint {
	var endpoints []ContractEndpoint

	pathPattern := regexp.MustCompile(`(?m)^  (/[^\s]+):\s*$`)
	pathLineIndices := pathPattern.FindAllStringSubmatchIndex(content, -1)

	for _, idx := range pathLineIndices {
		path := content[idx[2]:idx[3]]

		lineEnd := idx[1]
		nextPathOrEnd := len(content)
		for _, nextIdx := range pathLineIndices {
			if nextIdx[0] > lineEnd {
				nextPathOrEnd = nextIdx[0]
				break
			}
		}

		pathBlock := content[lineEnd:nextPathOrEnd]

		methodPattern := regexp.MustCompile(`(?m)^    (\w+):\s*$`)
		methodMatches := methodPattern.FindAllStringSubmatch(pathBlock, -1)

		for _, mMatch := range methodMatches {
			method := strings.ToUpper(mMatch[1])
			if !isValidHTTPMethod(method) {
				continue
			}

			methodLineStart := strings.Index(pathBlock, mMatch[0])
			methodBlockEnd := len(pathBlock)
			methodPattern2 := regexp.MustCompile(`(?m)^    \w+:\s*$`)
			nextMethods := methodPattern2.FindAllStringSubmatchIndex(pathBlock[methodLineStart+len(mMatch[0]):], -1)
			if len(nextMethods) > 0 {
				methodBlockEnd = methodLineStart + len(mMatch[0]) + nextMethods[0][0]
			}

			methodBlock := pathBlock[methodLineStart:methodBlockEnd]

			ep := ContractEndpoint{
				Path:   path,
				Method: method,
			}

			ep.Summary = extractMethodField(methodBlock, "summary")
			ep.HasRequest = strings.Contains(methodBlock, "requestBody:")
			ep.HasAuth = !strings.Contains(methodBlock, "security: []") && !strings.Contains(methodBlock, "security: []")

			respPattern := regexp.MustCompile(`'(\d{3})':|(\d{3}):`)
			respMatches := respPattern.FindAllStringSubmatch(methodBlock, -1)
			for _, rMatch := range respMatches {
				resp := rMatch[1]
				if resp == "" {
					resp = rMatch[2]
				}
				if resp != "" {
					ep.Responses = append(ep.Responses, resp)
				}
			}

			endpoints = append(endpoints, ep)
		}
	}

	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Path != endpoints[j].Path {
			return endpoints[i].Path < endpoints[j].Path
		}
		return methodOrder(endpoints[i].Method) < methodOrder(endpoints[j].Method)
	})

	return endpoints
}

func extractMethodField(methodBlock, field string) string {
	pattern := regexp.MustCompile(`(?m)^\s+` + field + `:\s*(.+)$`)
	match := pattern.FindStringSubmatch(methodBlock)
	if len(match) >= 2 {
		val := strings.TrimSpace(match[1])
		val = strings.Trim(val, "\"'")
		return val
	}
	return ""
}

func extractSchemas(content string) []ContractSchema {
	var schemas []ContractSchema

	schemaPattern := regexp.MustCompile(`(?m)^    (\w+):\s*\n\s+type:\s*object`)
	matches := schemaPattern.FindAllStringSubmatchIndex(content, -1)

	for _, idx := range matches {
		name := content[idx[2]:idx[3]]

		blockStart := idx[1]
		blockEnd := len(content)
		nextSchemaPattern := regexp.MustCompile(`(?m)^    \w+:\s*\n\s+type:\s*object`)
		nextMatches := nextSchemaPattern.FindAllStringSubmatchIndex(content[blockStart:], -1)
		if len(nextMatches) > 0 {
			blockEnd = blockStart + nextMatches[0][0]
		}

		block := content[blockStart:blockEnd]

		propPattern := regexp.MustCompile(`(?m)^\s+(\w+):\s*\n\s+type:\s*`)
		propMatches := propPattern.FindAllStringSubmatch(block, -1)
		var props []string
		for _, pm := range propMatches {
			props = append(props, pm[1])
		}

		schemas = append(schemas, ContractSchema{
			Name:       name,
			Properties: props,
		})
	}

	return schemas
}

func isValidHTTPMethod(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	}
	return false
}

func methodOrder(method string) int {
	switch method {
	case "GET":
		return 0
	case "POST":
		return 1
	case "PUT":
		return 2
	case "PATCH":
		return 3
	case "DELETE":
		return 4
	}
	return 5
}

func isPublicEndpoint(path string) bool {
	return strings.Contains(path, "/auth/") ||
		strings.Contains(path, "/login") ||
		strings.Contains(path, "/register") ||
		strings.Contains(path, "/health") ||
		strings.Contains(path, "/public/")
}

func ContractExists() bool {
	return ArtifactExists(ContractFile)
}

func FormatContractSpec(spec *ContractSpec) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Title:     %s\n", spec.Title))
	sb.WriteString(fmt.Sprintf("Version:   %s\n", spec.Version))
	sb.WriteString(fmt.Sprintf("Auth:      %v\n", spec.HasAuth))
	sb.WriteString(fmt.Sprintf("Endpoints: %d\n", spec.EndpointCount))
	sb.WriteString(fmt.Sprintf("Schemas:   %d\n", spec.SchemaCount))
	sb.WriteString("\n")

	if len(spec.Endpoints) > 0 {
		sb.WriteString("Endpoints:\n")
		sb.WriteString(fmt.Sprintf("  %-6s  %-30s  %-10s  %-8s  %-8s  %s\n",
			"METHOD", "PATH", "AUTH", "REQUEST", "RESP", "SUMMARY"))
		sb.WriteString(fmt.Sprintf("  %s  %s  %s  %s  %s  %s\n",
			strings.Repeat("-", 6), strings.Repeat("-", 30), strings.Repeat("-", 10),
			strings.Repeat("-", 8), strings.Repeat("-", 8), strings.Repeat("-", 30)))
		for _, ep := range spec.Endpoints {
			auth := "no"
			if ep.HasAuth {
				auth = "yes"
			}
			req := "no"
			if ep.HasRequest {
				req = "yes"
			}
			resps := strings.Join(ep.Responses, ",")
			sb.WriteString(fmt.Sprintf("  %-6s  %-30s  %-10s  %-8s  %-8s  %s\n",
				ep.Method, ep.Path, auth, req, resps, truncateStr(ep.Summary, 30)))
		}
	}

	if len(spec.Schemas) > 0 {
		sb.WriteString("\nSchemas:\n")
		for _, s := range spec.Schemas {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", s.Name, strings.Join(s.Properties, ", ")))
		}
	}

	return sb.String()
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
