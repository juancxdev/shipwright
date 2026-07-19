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
			"product/context.md",
			"product/assumptions.md",
			"product/open-questions.md",
		}

	case StateProductContextReady:
		return []string{
			"architecture/technology-options.md",
		}

	case StateTechnicalScopeDraft:
		return []string{
			"product/scope.md",
		}

	case StateScopeApproved:
		return []string{
			"project/project-charter.md",
			"project/project-plan.md",
			"project/risk-register.md",
		}

	case StateProjectPlanning:
		return []string{
			"project/delivery-plan.md",
		}

	case StateTechnicalDesign:
		return []string{
			"architecture/system-architecture.md",
			"backlog/epics.md",
			"backlog/user-stories.md",
			"backlog/frontend-tasks.md",
			"backlog/backend-tasks.md",
			"contracts/openapi.yaml",
			"sdd/proposal.md",
			"sdd/spec.md",
			"sdd/tasks.md",
		}

	case StateImplementation:
		return []string{
			"progress/frontend.md",
			"progress/backend.md",
		}

	case StateIntegration:
		return []string{
			"reports/contract-test-report.md",
			"reports/review-checklist.md",
		}

	case StateQASecurityReview:
		return []string{
			"reports/qa-report.md",
			"reports/security-review.md",
			"reports/review-checklist.md",
		}

	case StateUserAcceptance:
		return []string{
			"project/acceptance-report.md",
		}

	case StateChangeRequest:
		return []string{
			"project/change-management.md",
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
		"product/context.md",
		"product/assumptions.md",
		"product/open-questions.md",
		"product/scope.md",
		"architecture/technology-options.md",
		"architecture/system-architecture.md",
		"project/project-charter.md",
		"project/project-plan.md",
		"project/risk-register.md",
		"project/delivery-plan.md",
		"project/change-management.md",
		"project/acceptance-report.md",
		"contracts/openapi.yaml",
		"backlog/epics.md",
		"backlog/user-stories.md",
		"backlog/frontend-tasks.md",
		"backlog/backend-tasks.md",
		"sdd/proposal.md",
		"sdd/spec.md",
		"sdd/tasks.md",
		"progress/frontend.md",
		"progress/backend.md",
		"reports/contract-test-report.md",
		"reports/qa-report.md",
		"reports/security-review.md",
		"reports/review-checklist.md",
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
