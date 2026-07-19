package harness

import (
	"fmt"
	"strings"
)

type StateValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
	Repaired []string
}

func ValidateState(s *State) *StateValidationResult {
	result := &StateValidationResult{Valid: true}
	if s == nil {
		return &StateValidationResult{Valid: false, Errors: []string{"state is nil"}}
	}

	if strings.TrimSpace(s.ProjectName) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "project_name is required")
	}
	if strings.TrimSpace(s.ProjectID) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "project_id is required")
	}
	if !StateExists(s.CurrentPhase) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("invalid current_phase: %s", s.CurrentPhase))
	}
	if !validStatus(s.Status) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("invalid status: %s", s.Status))
	}
	if s.Status == StatusBlocked && strings.TrimSpace(s.BlockReason) == "" {
		result.Warnings = append(result.Warnings, "blocked state has empty block_reason")
	}
	if s.CreatedAt == "" {
		result.Warnings = append(result.Warnings, "created_at is empty")
	}
	if s.UpdatedAt == "" {
		result.Warnings = append(result.Warnings, "updated_at is empty")
	}

	for _, gate := range AllGates() {
		if _, ok := s.Approvals[gate]; !ok {
			result.Warnings = append(result.Warnings, fmt.Sprintf("approval gate missing: %s", gate))
		}
	}

	return result
}

func RepairState(s *State) *StateValidationResult {
	result := &StateValidationResult{Valid: true}
	if s == nil {
		return &StateValidationResult{Valid: false, Errors: []string{"state is nil"}}
	}

	if strings.TrimSpace(s.ProjectName) == "" {
		s.ProjectName = "Recovered Project"
		result.Repaired = append(result.Repaired, "project_name set to Recovered Project")
	}
	if strings.TrimSpace(s.ProjectID) == "" {
		s.ProjectID = generateProjectID(s.ProjectName)
		result.Repaired = append(result.Repaired, "project_id regenerated")
	}
	if !StateExists(s.CurrentPhase) {
		s.CurrentPhase = StateIntake
		s.SetBlocked("state.json had invalid current_phase; recovered to INTAKE")
		result.Repaired = append(result.Repaired, "current_phase recovered to INTAKE")
	}
	if !validStatus(s.Status) {
		s.Status = StatusReady
		result.Repaired = append(result.Repaired, "status recovered to ready")
	}
	if s.Approvals == nil {
		s.Approvals = make(ApprovalMap)
		result.Repaired = append(result.Repaired, "approvals map initialized")
	}
	for _, gate := range AllGates() {
		if _, ok := s.Approvals[gate]; !ok {
			s.Approvals[gate] = false
			result.Repaired = append(result.Repaired, "approval gate initialized: "+gate)
		}
	}
	if s.CreatedAt == "" {
		s.CreatedAt = NowISO()
		result.Repaired = append(result.Repaired, "created_at initialized")
	}
	if s.UpdatedAt == "" {
		s.UpdatedAt = NowISO()
		result.Repaired = append(result.Repaired, "updated_at initialized")
	}

	validation := ValidateState(s)
	result.Valid = validation.Valid
	result.Errors = append(result.Errors, validation.Errors...)
	result.Warnings = append(result.Warnings, validation.Warnings...)
	return result
}

func AllGates() []string {
	return []string{
		GateScope,
		GateUXDesign,
		GateTechnicalPlan,
		GateTechLeadReview,
		GateFinalAcceptance,
	}
}

func validStatus(status string) bool {
	switch status {
	case StatusBlocked, StatusReady, StatusClosed:
		return true
	default:
		return false
	}
}
