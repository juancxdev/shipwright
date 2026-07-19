package harness

import (
	"fmt"
	"strings"
)

type ContractValidation struct {
	IsValid   bool
	Errors    []string
	Warnings  []string
	Endpoints int
	Schemas   int
}

func ValidateContract(path string) *ContractValidation {
	result := &ContractValidation{}

	if !ArtifactExists(path) {
		result.Errors = append(result.Errors, fmt.Sprintf("%s does not exist", path))
		return result
	}

	parseResult := ParseContract(path)
	result.Errors = parseResult.Errors
	result.Warnings = parseResult.Warnings
	result.IsValid = parseResult.IsValid

	if parseResult.Spec != nil {
		result.Endpoints = parseResult.Spec.EndpointCount
		result.Schemas = parseResult.Spec.SchemaCount
	}

	return result
}

func GenerateFrontendTasks(spec *ContractSpec, projectName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Frontend Tasks\n\n"))
	sb.WriteString(fmt.Sprintf("**Project:** %s\n", projectName))
	sb.WriteString(fmt.Sprintf("**Generated from:** %s\n", ContractFile))
	sb.WriteString(fmt.Sprintf("**Contract endpoints:** %d\n\n", spec.EndpointCount))

	sb.WriteString("## Contract-First Rules\n\n")
	sb.WriteString("- All tasks MUST reference endpoints from contracts/openapi.yaml\n")
	sb.WriteString("- Mock mode is MANDATORY — every data-fetching task must have mock + HTTP mode\n")
	sb.WriteString("- Mocks MUST derive from contract response schemas\n")
	sb.WriteString("- HTTP adapter MUST be separate from mock adapter\n")
	sb.WriteString("- NEVER delete mocks — mock mode must coexist with real API mode\n")
	sb.WriteString("- NEVER invent endpoints not in the contract\n\n")

	sb.WriteString("## Task List\n\n")

	taskNum := 1

	if spec.HasAuth {
		sb.WriteString(fmt.Sprintf("### %d. Auth adapter (mock + HTTP)\n\n", taskNum))
		sb.WriteString("- [ ] Create auth service with mock mode (returns fake token)\n")
		sb.WriteString("- [ ] Create auth service with HTTP mode (calls real auth endpoint)\n")
		sb.WriteString("- [ ] Add toggle between mock and HTTP mode\n")
		sb.WriteString("- [ ] Store token in secure storage\n\n")
		taskNum++
	}

	for _, ep := range spec.Endpoints {
		sb.WriteString(fmt.Sprintf("### %d. %s %s — %s\n\n", taskNum, ep.Method, ep.Path, ep.Summary))

		resourceName := extractResourceName(ep.Path)
		actionVerb := verbForMethod(ep.Method)

		sb.WriteString(fmt.Sprintf("- [ ] Create %s component for %s\n", resourceName, actionVerb))
		sb.WriteString(fmt.Sprintf("- [ ] Create data-fetching hook/service for `%s %s`\n", ep.Method, ep.Path))

		if ep.HasRequest {
			sb.WriteString("- [ ] Create request form/body matching contract request schema\n")
		}

		sb.WriteString("- [ ] Implement mock mode (return fake data matching response schema)\n")
		sb.WriteString("- [ ] Implement HTTP mode (call real API)\n")
		sb.WriteString("- [ ] Add mode toggle (mock ↔ HTTP)\n")

		sb.WriteString("- [ ] Handle states: loading, empty, error, success\n")
		sb.WriteString(fmt.Sprintf("- [ ] Handle error responses: %s\n", strings.Join(filterErrorResponses(ep.Responses), ", ")))

		if ep.HasAuth {
			sb.WriteString("- [ ] Attach auth token to request\n")
		}

		sb.WriteString(fmt.Sprintf("- [ ] Write test for %s component\n\n", resourceName))
		taskNum++
	}

	if len(spec.Schemas) > 0 {
		sb.WriteString(fmt.Sprintf("### %d. TypeScript types from contract schemas\n\n", taskNum))
		sb.WriteString("Generate TypeScript interfaces from contract schemas:\n\n")
		for _, s := range spec.Schemas {
			sb.WriteString(fmt.Sprintf("- [ ] `%s` interface with: %s\n", s.Name, strings.Join(s.Properties, ", ")))
		}
		sb.WriteString("\n")
		taskNum++
	}

	sb.WriteString(fmt.Sprintf("### %d. Mode integration test\n\n", taskNum))
	sb.WriteString("- [ ] Verify all components can run in mock mode\n")
	sb.WriteString("- [ ] Verify all components can run in HTTP mode\n")
	sb.WriteString("- [ ] Verify mode toggle works without page reload\n")
	sb.WriteString("- [ ] Verify no mocks were deleted\n\n")

	sb.WriteString("## Contract Reference\n\n")
	sb.WriteString("| Method | Path | Summary |\n")
	sb.WriteString("|--------|------|---------|\n")
	for _, ep := range spec.Endpoints {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", ep.Method, ep.Path, ep.Summary))
	}

	return sb.String()
}

func GenerateBackendTasks(spec *ContractSpec, projectName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Backend Tasks\n\n"))
	sb.WriteString(fmt.Sprintf("**Project:** %s\n", projectName))
	sb.WriteString(fmt.Sprintf("**Generated from:** %s\n", ContractFile))
	sb.WriteString(fmt.Sprintf("**Contract endpoints:** %d\n\n", spec.EndpointCount))

	sb.WriteString("## Contract-First Rules\n\n")
	sb.WriteString("- All tasks MUST implement endpoints from contracts/openapi.yaml EXACTLY\n")
	sb.WriteString("- API MUST match contract: path, method, request schema, response schema\n")
	sb.WriteString("- Error responses MUST match contract format (400, 401, 500 minimum)\n")
	sb.WriteString("- NEVER break the contract without a change request\n")
	sb.WriteString("- If the contract is wrong, STOP and request a change request\n")
	sb.WriteString("- Every endpoint MUST have consistent error responses\n")
	sb.WriteString("- Every endpoint MUST have tests (domain + API)\n\n")

	sb.WriteString("## Task List\n\n")

	taskNum := 1

	if spec.HasAuth {
		sb.WriteString(fmt.Sprintf("### %d. Auth middleware\n\n", taskNum))
		sb.WriteString("- [ ] Implement authentication middleware matching contract security scheme\n")
		sb.WriteString("- [ ] Implement token validation\n")
		sb.WriteString("- [ ] Implement authorization checks\n")
		sb.WriteString("- [ ] Write tests for auth middleware\n\n")
		taskNum++
	}

	if len(spec.Schemas) > 0 {
		sb.WriteString(fmt.Sprintf("### %d. Data models from contract schemas\n\n", taskNum))
		sb.WriteString("Implement data models matching contract schemas:\n\n")
		for _, s := range spec.Schemas {
			sb.WriteString(fmt.Sprintf("- [ ] `%s` model with fields: %s\n", s.Name, strings.Join(s.Properties, ", ")))
		}
		sb.WriteString("- [ ] Write tests for each model\n\n")
		taskNum++
	}

	for _, ep := range spec.Endpoints {
		sb.WriteString(fmt.Sprintf("### %d. %s %s — %s\n\n", taskNum, ep.Method, ep.Path, ep.Summary))

		resourceName := extractResourceName(ep.Path)

		sb.WriteString(fmt.Sprintf("- [ ] Implement `%s %s` endpoint\n", ep.Method, ep.Path))
		sb.WriteString(fmt.Sprintf("- [ ] Implement %s domain logic\n", resourceName))

		if ep.HasRequest {
			sb.WriteString("- [ ] Validate request body against contract schema\n")
		}

		sb.WriteString(fmt.Sprintf("- [ ] Return response matching contract schema (%s)\n",
			strings.Join(filterSuccessResponses(ep.Responses), ", ")))
		sb.WriteString(fmt.Sprintf("- [ ] Return error responses: %s\n",
			strings.Join(filterErrorResponses(ep.Responses), ", ")))

		if ep.HasAuth {
			sb.WriteString("- [ ] Enforce authentication on this endpoint\n")
		}

		sb.WriteString(fmt.Sprintf("- [ ] Write domain test for %s\n", resourceName))
		sb.WriteString(fmt.Sprintf("- [ ] Write API test for `%s %s`\n\n", ep.Method, ep.Path))
		taskNum++
	}

	sb.WriteString(fmt.Sprintf("### %d. Contract compliance test\n\n", taskNum))
	sb.WriteString("- [ ] Verify all endpoints match contracts/openapi.yaml\n")
	sb.WriteString("- [ ] Verify all error responses are consistent\n")
	sb.WriteString("- [ ] Verify all schemas match contract definitions\n")
	sb.WriteString("- [ ] Run contract tests against live API\n\n")

	sb.WriteString("## Contract Reference\n\n")
	sb.WriteString("| Method | Path | Summary |\n")
	sb.WriteString("|--------|------|---------|\n")
	for _, ep := range spec.Endpoints {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", ep.Method, ep.Path, ep.Summary))
	}

	return sb.String()
}

func extractResourceName(path string) string {
	segments := splitPath(path)
	if len(segments) == 0 {
		return "root"
	}

	for i := len(segments) - 1; i >= 0; i-- {
		seg := segments[i]
		if !strings.HasPrefix(seg, "{") {
			return toPascalCase(seg)
		}
	}

	if len(segments) > 0 {
		return toPascalCase(segments[0])
	}
	return "root"
}

func splitPath(path string) []string {
	var result []string
	for _, seg := range strings.Split(path, "/") {
		if seg != "" {
			result = append(result, seg)
		}
	}
	return result
}

func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	words := strings.Split(s, "-")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, "")
}

func verbForMethod(method string) string {
	switch method {
	case "GET":
		return "fetching"
	case "POST":
		return "creating"
	case "PUT":
		return "updating"
	case "PATCH":
		return "patching"
	case "DELETE":
		return "deleting"
	}
	return "handling"
}

func filterErrorResponses(responses []string) []string {
	var result []string
	for _, r := range responses {
		if strings.HasPrefix(r, "4") || strings.HasPrefix(r, "5") {
			result = append(result, r)
		}
	}
	if len(result) == 0 {
		return []string{"400", "401", "500"}
	}
	return result
}

func filterSuccessResponses(responses []string) []string {
	var result []string
	for _, r := range responses {
		if strings.HasPrefix(r, "2") {
			result = append(result, r)
		}
	}
	if len(result) == 0 {
		return []string{"200"}
	}
	return result
}

func GenerateContractTasks(spec *ContractSpec, projectName string) (feTasks, beTasks string) {
	feTasks = GenerateFrontendTasks(spec, projectName)
	beTasks = GenerateBackendTasks(spec, projectName)
	return
}
