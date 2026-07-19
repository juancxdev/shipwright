package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const TDDPolicyJSON = ".harness/tdd-policy.json"
const TDDPolicyMarkdown = ".harness/tdd-policy.md"
const TDDPolicyVersion = "1"
const TDDReportFile = "reports/tdd-report.md"

const (
	TDDModeStrict    = "strict"
	TDDModeSuggested = "suggested"
	TDDModeNone      = "none"
)

type TDDPolicy struct {
	Version          string   `json:"version"`
	GeneratedAt      string   `json:"generated_at"`
	Mode             string   `json:"mode"`
	Supported        bool     `json:"supported"`
	TestCommand      string   `json:"test_command,omitempty"`
	Source           string   `json:"source"`
	RequiredEvidence []string `json:"required_evidence,omitempty"`
	Warnings         []string `json:"warnings,omitempty"`
}

type TDDAssessment struct {
	PolicyPresent       bool     `json:"policy_present"`
	Mode                string   `json:"mode"`
	Supported           bool     `json:"supported"`
	TestCommand         string   `json:"test_command,omitempty"`
	HasFrontendEvidence bool     `json:"has_frontend_evidence"`
	HasBackendEvidence  bool     `json:"has_backend_evidence"`
	HasReportEvidence   bool     `json:"has_report_evidence"`
	Issues              []string `json:"issues,omitempty"`
	Warnings            []string `json:"warnings,omitempty"`
}

func RefreshTDDPolicy() (*TDDPolicy, error) {
	profile, err := LoadProjectProfile()
	if err != nil {
		return nil, err
	}
	return RefreshTDDPolicyFromProfile(profile)
}

func RefreshTDDPolicyFromProfile(profile *ProjectProfile) (*TDDPolicy, error) {
	policy := BuildTDDPolicy(profile)
	if err := SaveTDDPolicy(policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func BuildTDDPolicy(profile *ProjectProfile) *TDDPolicy {
	policy := &TDDPolicy{
		Version:     TDDPolicyVersion,
		GeneratedAt: NowISO(),
		Mode:        TDDModeNone,
		Source:      ProjectProfileJSON,
	}
	if profile == nil {
		policy.Source = "default"
		policy.Warnings = append(policy.Warnings, "project profile missing; TDD mode defaults to none")
		return policy
	}

	policy.Mode = normalizeTDDMode(profile.TDD.RecommendedMode)
	policy.Supported = profile.TDD.Supported
	if len(profile.Commands.Test) > 0 {
		policy.TestCommand = profile.Commands.Test[0].Command
	}

	switch policy.Mode {
	case TDDModeStrict:
		policy.RequiredEvidence = []string{
			"progress/frontend.md must include TDD/Test evidence, or reports/tdd-report.md must cover frontend evidence",
			"progress/backend.md must include TDD/Test evidence, or reports/tdd-report.md must cover backend evidence",
			"Evidence must mention the detected test command when available",
		}
	case TDDModeSuggested:
		policy.RequiredEvidence = []string{"Record test evidence when implementation touches behavior"}
	case TDDModeNone:
		policy.RequiredEvidence = []string{"No strict TDD gate; record manual evidence if tests are unavailable"}
	}
	return policy
}

func SaveTDDPolicy(policy *TDDPolicy) error {
	if policy == nil {
		return fmt.Errorf("TDD policy is nil")
	}
	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return err
	}
	if err := WriteFile(TDDPolicyJSON, string(data)+"\n"); err != nil {
		return err
	}
	return WriteFile(TDDPolicyMarkdown, RenderTDDPolicyMarkdown(policy))
}

func LoadTDDPolicy() (*TDDPolicy, error) {
	data, err := os.ReadFile(TDDPolicyJSON)
	if err != nil {
		return nil, err
	}
	var policy TDDPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, err
	}
	policy.Mode = normalizeTDDMode(policy.Mode)
	return &policy, nil
}

func AssessTDDCompliance() TDDAssessment {
	policy, present := effectiveTDDPolicy()
	assessment := TDDAssessment{PolicyPresent: present}
	if policy == nil {
		assessment.Mode = TDDModeNone
		assessment.Warnings = append(assessment.Warnings, "TDD policy not found and project profile could not be loaded")
		return assessment
	}

	assessment.Mode = normalizeTDDMode(policy.Mode)
	assessment.Supported = policy.Supported
	assessment.TestCommand = policy.TestCommand
	assessment.HasFrontendEvidence = fileHasTDDEvidence("progress/frontend.md", policy)
	assessment.HasBackendEvidence = fileHasTDDEvidence("progress/backend.md", policy)
	assessment.HasReportEvidence = fileHasTDDEvidence(TDDReportFile, policy)
	if !present {
		assessment.Warnings = append(assessment.Warnings, "TDD policy file is missing; inferred mode from project profile")
	}
	assessment.Warnings = append(assessment.Warnings, policy.Warnings...)

	if assessment.Mode == TDDModeStrict {
		if !assessment.HasReportEvidence && !assessment.HasFrontendEvidence {
			assessment.Issues = append(assessment.Issues, "strict TDD requires frontend test evidence in progress/frontend.md or reports/tdd-report.md")
		}
		if !assessment.HasReportEvidence && !assessment.HasBackendEvidence {
			assessment.Issues = append(assessment.Issues, "strict TDD requires backend test evidence in progress/backend.md or reports/tdd-report.md")
		}
	}
	return assessment
}

func TDDBlockReason() string {
	assessment := AssessTDDCompliance()
	if assessment.Mode != TDDModeStrict || len(assessment.Issues) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Strict TDD gate blocked IMPLEMENTATION → INTEGRATION.\n")
	sb.WriteString("Shipwright detected test capability and requires real TDD/test evidence before integration.\n\n")
	sb.WriteString("Missing evidence:\n")
	for _, issue := range assessment.Issues {
		sb.WriteString("  - ")
		sb.WriteString(issue)
		sb.WriteString("\n")
	}
	sb.WriteString("\nRun: shipwright tdd status\n")
	sb.WriteString("Then update progress/frontend.md, progress/backend.md, or reports/tdd-report.md with the executed test evidence.")
	return sb.String()
}

func RenderTDDPolicyMarkdown(policy *TDDPolicy) string {
	var sb strings.Builder
	sb.WriteString("# TDD Policy\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n", policy.GeneratedAt))
	sb.WriteString(fmt.Sprintf("**Mode:** `%s`\n", normalizeTDDMode(policy.Mode)))
	sb.WriteString(fmt.Sprintf("**Supported:** `%s`\n", yesNo(policy.Supported)))
	if policy.TestCommand != "" {
		sb.WriteString(fmt.Sprintf("**Detected test command:** `%s`\n", policy.TestCommand))
	} else {
		sb.WriteString("**Detected test command:** _none_\n")
	}
	sb.WriteString(fmt.Sprintf("**Source:** `%s`\n\n", policy.Source))

	sb.WriteString("## Gate behavior\n\n")
	switch normalizeTDDMode(policy.Mode) {
	case TDDModeStrict:
		sb.WriteString("Shipwright blocks `IMPLEMENTATION → INTEGRATION` until implementation progress contains real TDD/test evidence.\n\n")
	case TDDModeSuggested:
		sb.WriteString("Shipwright recommends test evidence, but does not block integration only for TDD evidence.\n\n")
	default:
		sb.WriteString("Shipwright does not enforce strict TDD because no reliable test capability was detected.\n\n")
	}

	sb.WriteString("## Required evidence\n\n")
	for _, item := range policy.RequiredEvidence {
		sb.WriteString(fmt.Sprintf("- %s\n", item))
	}
	if len(policy.RequiredEvidence) == 0 {
		sb.WriteString("- No explicit evidence configured.\n")
	}

	if len(policy.Warnings) > 0 {
		sb.WriteString("\n## Warnings\n\n")
		for _, warning := range policy.Warnings {
			sb.WriteString(fmt.Sprintf("- %s\n", warning))
		}
	}

	sb.WriteString("\n## How agents must use this\n\n")
	sb.WriteString("- Read this policy before implementation work.\n")
	sb.WriteString("- If mode is `strict`, write tests before or alongside behavior changes and record evidence.\n")
	sb.WriteString("- Do not claim done unless progress files or `reports/tdd-report.md` contain executed test evidence.\n")
	return sb.String()
}

func FormatTDDAssessment(assessment TDDAssessment) string {
	var sb strings.Builder
	sb.WriteString("Shipwright — TDD Status\n")
	sb.WriteString(strings.Repeat("=", 30))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("Policy file:      %s\n", yesNo(assessment.PolicyPresent)))
	sb.WriteString(fmt.Sprintf("Mode:             %s\n", assessment.Mode))
	sb.WriteString(fmt.Sprintf("Supported:        %s\n", yesNo(assessment.Supported)))
	if assessment.TestCommand != "" {
		sb.WriteString(fmt.Sprintf("Test command:     %s\n", assessment.TestCommand))
	} else {
		sb.WriteString("Test command:     not detected\n")
	}
	sb.WriteString("\nEvidence:\n")
	sb.WriteString(fmt.Sprintf("  frontend:       %s\n", checkMark(assessment.HasFrontendEvidence)))
	sb.WriteString(fmt.Sprintf("  backend:        %s\n", checkMark(assessment.HasBackendEvidence)))
	sb.WriteString(fmt.Sprintf("  tdd report:     %s\n", checkMark(assessment.HasReportEvidence)))
	if len(assessment.Issues) > 0 {
		sb.WriteString("\nBlocking issues:\n")
		for _, issue := range assessment.Issues {
			sb.WriteString("  - ")
			sb.WriteString(issue)
			sb.WriteString("\n")
		}
	}
	if len(assessment.Warnings) > 0 {
		sb.WriteString("\nWarnings:\n")
		for _, warning := range assessment.Warnings {
			sb.WriteString("  - ")
			sb.WriteString(warning)
			sb.WriteString("\n")
		}
	}
	if len(assessment.Issues) == 0 {
		sb.WriteString("\n✓ TDD gate is satisfied for the current policy.\n")
	} else {
		sb.WriteString("\n✗ TDD gate is blocked.\n")
	}
	return sb.String()
}

func effectiveTDDPolicy() (*TDDPolicy, bool) {
	policy, err := LoadTDDPolicy()
	if err == nil {
		return policy, true
	}
	profile, err := LoadProjectProfile()
	if err != nil {
		return nil, false
	}
	return BuildTDDPolicy(profile), false
}

func fileHasTDDEvidence(path string, policy *TDDPolicy) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	content := strings.ToLower(string(data))
	markers := []string{
		"tdd evidence:",
		"test evidence:",
		"red:",
		"green:",
		"refactor:",
		"tests passing",
		"all passing",
		"passing tests",
		"test passed",
		"tests passed",
		"go test",
		"npm test",
		"pnpm test",
		"yarn test",
		"bun test",
		"pytest",
		"cargo test",
		"mvn test",
		"gradle test",
	}
	if policy != nil && policy.TestCommand != "" {
		markers = append(markers, strings.ToLower(policy.TestCommand))
	}
	for _, marker := range markers {
		if strings.Contains(content, marker) {
			return true
		}
	}
	return false
}

func normalizeTDDMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case TDDModeStrict:
		return TDDModeStrict
	case TDDModeSuggested:
		return TDDModeSuggested
	default:
		return TDDModeNone
	}
}

func checkMark(value bool) string {
	if value {
		return "✓"
	}
	return "✗"
}
