package harness

import (
	"fmt"
	"strings"
)

type ScaffoldResult struct {
	Generated []string
	Skipped   []string
	Errors    []string
}

func ArtifactsForPhase(phase string, requiresUI *bool) []string {
	switch phase {
	case StateDiscovery:
		return []string{
			".harness/artifacts/product/context.md",
			".harness/artifacts/product/assumptions.md",
			".harness/artifacts/product/open-questions.md",
		}

	case StateProductContextReady:
		return []string{
			".harness/artifacts/architecture/technology-options.md",
		}

	case StateTechnicalScopeDraft:
		return []string{
			".harness/artifacts/product/scope.md",
		}

	case StateScopeApproved:
		return []string{
			".harness/artifacts/project/project-charter.md",
			".harness/artifacts/project/project-plan.md",
			".harness/artifacts/project/risk-register.md",
		}

	case StateProjectPlanning:
		return []string{
			".harness/artifacts/project/delivery-plan.md",
		}

	case StateTechnicalDesign:
		return []string{
			".harness/artifacts/architecture/system-architecture.md",
			".harness/artifacts/backlog/epics.md",
			".harness/artifacts/backlog/user-stories.md",
			".harness/artifacts/backlog/frontend-tasks.md",
			".harness/artifacts/backlog/backend-tasks.md",
			".harness/artifacts/contracts/openapi.yaml",
			".harness/artifacts/sdd/proposal.md",
			".harness/artifacts/sdd/spec.md",
			".harness/artifacts/sdd/tasks.md",
		}

	case StateImplementation:
		return []string{
			".harness/artifacts/progress/frontend.md",
			".harness/artifacts/progress/backend.md",
		}

	case StateIntegration:
		return []string{
			".harness/artifacts/reports/contract-test-report.md",
			".harness/artifacts/reports/review-checklist.md",
		}

	case StateQASecurityReview:
		return []string{
			".harness/artifacts/reports/qa-report.md",
			".harness/artifacts/reports/security-review.md",
			".harness/artifacts/reports/review-checklist.md",
		}

	case StateUserAcceptance:
		return []string{
			".harness/artifacts/project/acceptance-report.md",
		}

	case StateChangeRequest:
		return []string{
			".harness/artifacts/project/change-management.md",
		}

	default:
		return nil
	}
}

func ScaffoldPhase(s *State) *ScaffoldResult {
	data := TemplateDataFromState(s)
	artifacts := ArtifactsForPhase(s.CurrentPhase, s.RequiresUI)

	result := &ScaffoldResult{}

	for _, path := range artifacts {
		if ArtifactExists(path) {
			result.Skipped = append(result.Skipped, path)
			continue
		}

		content, err := GenerateArtifact(path, data)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", path, err))
			continue
		}

		if err := WriteFile(path, content); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", path, err))
			continue
		}

		result.Generated = append(result.Generated, path)
	}

	return result
}

func ScaffoldArtifact(s *State, path string) error {
	if ArtifactExists(path) {
		return fmt.Errorf("file already exists: %s", path)
	}

	data := TemplateDataFromState(s)
	content, err := GenerateArtifact(path, data)
	if err != nil {
		return err
	}

	return WriteFile(path, content)
}

func ListScaffoldableArtifacts() []string {
	return []string{
		".harness/artifacts/product/context.md",
		".harness/artifacts/product/assumptions.md",
		".harness/artifacts/product/open-questions.md",
		".harness/artifacts/product/scope.md",
		".harness/artifacts/architecture/technology-options.md",
		".harness/artifacts/architecture/system-architecture.md",
		".harness/artifacts/project/project-charter.md",
		".harness/artifacts/project/project-plan.md",
		".harness/artifacts/project/risk-register.md",
		".harness/artifacts/project/delivery-plan.md",
		".harness/artifacts/project/change-management.md",
		".harness/artifacts/project/acceptance-report.md",
		".harness/artifacts/contracts/openapi.yaml",
		".harness/artifacts/backlog/epics.md",
		".harness/artifacts/backlog/user-stories.md",
		".harness/artifacts/backlog/frontend-tasks.md",
		".harness/artifacts/backlog/backend-tasks.md",
		".harness/artifacts/sdd/proposal.md",
		".harness/artifacts/sdd/spec.md",
		".harness/artifacts/sdd/tasks.md",
		".harness/artifacts/progress/frontend.md",
		".harness/artifacts/progress/backend.md",
		".harness/artifacts/reports/contract-test-report.md",
		".harness/artifacts/reports/qa-report.md",
		".harness/artifacts/reports/security-review.md",
		".harness/artifacts/reports/review-checklist.md",
	}
}

func FormatScaffoldResult(r *ScaffoldResult) string {
	var sb strings.Builder

	if len(r.Generated) > 0 {
		sb.WriteString("Generated:\n")
		for _, f := range r.Generated {
			sb.WriteString(fmt.Sprintf("  ✓ %s\n", f))
		}
	}

	if len(r.Skipped) > 0 {
		sb.WriteString("Skipped (already exist):\n")
		for _, f := range r.Skipped {
			sb.WriteString(fmt.Sprintf("  → %s\n", f))
		}
	}

	if len(r.Errors) > 0 {
		sb.WriteString("Errors:\n")
		for _, e := range r.Errors {
			sb.WriteString(fmt.Sprintf("  ✗ %s\n", e))
		}
	}

	if len(r.Generated) == 0 && len(r.Skipped) == 0 && len(r.Errors) == 0 {
		sb.WriteString("Nothing to scaffold for current phase.\n")
	}

	return sb.String()
}
