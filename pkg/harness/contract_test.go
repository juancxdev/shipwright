package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validOpenAPI = `openapi: 3.0.3
info:
  title: Billing API
  version: 1.0.0
paths:
  /invoices:
    get:
      summary: List invoices
      responses:
        '200':
          description: Success
        '400':
          description: Bad request
    post:
      summary: Create invoice
      requestBody:
        required: true
      responses:
        '201':
          description: Created
        '400':
          description: Bad request
  /invoices/{id}:
    get:
      summary: Get invoice
      responses:
        '200':
          description: Success
        '404':
          description: Not found
components:
  schemas:
    Invoice:
      type: object
      properties:
        id:
          type: string
        total:
          type: number
`

func chdirTemp(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
	return dir
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestParseContractParsesEndpointsAndSchemas(t *testing.T) {
	chdirTemp(t)
	writeTestFile(t, ContractFile, validOpenAPI)

	result := ParseContract(ContractFile)
	if !result.IsValid {
		t.Fatalf("expected valid contract, got errors: %v", result.Errors)
	}
	if result.Spec == nil {
		t.Fatal("expected parsed spec")
	}
	if got, want := result.Spec.EndpointCount, 3; got != want {
		t.Fatalf("EndpointCount = %d, want %d", got, want)
	}
	if got, want := result.Spec.SchemaCount, 1; got != want {
		t.Fatalf("SchemaCount = %d, want %d", got, want)
	}
	if got, want := result.Spec.Title, "Billing API"; got != want {
		t.Fatalf("Title = %q, want %q", got, want)
	}
}

func TestValidateContractFailsWhenMissing(t *testing.T) {
	chdirTemp(t)

	result := ValidateContract(ContractFile)
	if result.IsValid {
		t.Fatal("expected missing contract to be invalid")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected missing contract error")
	}
}

func TestGenerateContractTasksIncludesContractFirstRules(t *testing.T) {
	chdirTemp(t)
	writeTestFile(t, ContractFile, validOpenAPI)
	parseResult := ParseContract(ContractFile)
	if !parseResult.IsValid {
		t.Fatalf("expected valid contract, got errors: %v", parseResult.Errors)
	}

	frontend, backend := GenerateContractTasks(parseResult.Spec, "Billing")

	assertContains(t, frontend, "Mock mode is MANDATORY")
	assertContains(t, frontend, "HTTP adapter MUST be separate")
	assertContains(t, frontend, "GET /invoices")
	assertContains(t, frontend, "POST /invoices")
	assertContains(t, backend, "API MUST match contract")
	assertContains(t, backend, "NEVER break the contract without a change request")
	assertContains(t, backend, "Invoice")
}

func TestCheckMockComplianceRequiresMockMode(t *testing.T) {
	chdirTemp(t)
	writeTestFile(t, ContractFile, validOpenAPI)
	writeTestFile(t, "progress/frontend.md", "# Frontend Progress\n\nImplemented HTTP mode for /invoices.\n")

	spec := ParseContract(ContractFile).Spec
	result := CheckMockCompliance(spec)

	if len(result.Issues) == 0 {
		t.Fatal("expected mock compliance issue")
	}
	assertContains(t, strings.Join(result.Issues, "\n"), "mock mode")
}

func TestCheckBackendComplianceRequiresErrorFormat(t *testing.T) {
	chdirTemp(t)
	writeTestFile(t, ContractFile, validOpenAPI)
	writeTestFile(t, "progress/backend.md", "# Backend Progress\n\nImplemented GET /invoices and POST /invoices using Invoice model.\n")

	spec := ParseContract(ContractFile).Spec
	result := CheckBackendCompliance(spec)

	if len(result.Issues) == 0 {
		t.Fatal("expected backend compliance issue")
	}
	assertContains(t, strings.Join(result.Issues, "\n"), "consistent error response format")
}

func assertContains(t *testing.T, text, expected string) {
	t.Helper()

	if !strings.Contains(text, expected) {
		t.Fatalf("expected %q to contain %q", text, expected)
	}
}
