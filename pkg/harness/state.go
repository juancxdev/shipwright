package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const StateDir = ".harness"
const StateFile = ".harness/state.json"

type ApprovalMap map[string]bool

type State struct {
	ProjectID           string      `json:"project_id"`
	ProjectName         string      `json:"project_name"`
	InitialRequest      string      `json:"initial_request,omitempty"`
	CurrentPhase        string      `json:"current_phase"`
	Status              string      `json:"status"`
	BlockReason         string      `json:"block_reason,omitempty"`
	Approvals           ApprovalMap `json:"approvals"`
	RequiresUI          *bool       `json:"requires_ui,omitempty"`
	ActiveChangeRequest *string     `json:"active_change_request,omitempty"`
	CreatedAt           string      `json:"created_at"`
	UpdatedAt           string      `json:"updated_at"`
}

const (
	StatusBlocked = "blocked"
	StatusReady   = "ready"
	StatusClosed  = "closed"
)

func nowISO() string {
	return NowISO()
}

func NowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func NewState(projectName string) *State {
	return &State{
		ProjectID:    generateProjectID(projectName),
		ProjectName:  projectName,
		CurrentPhase: StateIntake,
		Status:       StatusReady,
		Approvals: ApprovalMap{
			GateScope:           false,
			GateUXDesign:        false,
			GateTechnicalPlan:   false,
			GateTechLeadReview:  false,
			GateFinalAcceptance: false,
		},
		CreatedAt: nowISO(),
		UpdatedAt: nowISO(),
	}
}

func generateProjectID(name string) string {
	if name == "" {
		return "unnamed-project"
	}
	result := make([]byte, 0, len(name))
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result = append(result, byte(r))
		} else if r >= 'A' && r <= 'Z' {
			result = append(result, byte(r+32))
		} else if r == ' ' || r == '_' || r == '-' {
			result = append(result, '-')
		}
	}
	if len(result) == 0 {
		return "unnamed-project"
	}
	return string(result)
}

func HarnessInitialized() bool {
	info, err := os.Stat(StateFile)
	return err == nil && !info.IsDir()
}

func LoadState() (*State, error) {
	data, err := os.ReadFile(StateFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", StateFile, err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", StateFile, err)
	}

	repair := RepairState(&s)
	if !repair.Valid {
		return nil, fmt.Errorf("invalid %s: %s", StateFile, strings.Join(repair.Errors, "; "))
	}
	if len(repair.Repaired) > 0 {
		if err := s.Save(); err != nil {
			return nil, fmt.Errorf("cannot save repaired %s: %w", StateFile, err)
		}
	}
	return &s, nil
}

func (s *State) Save() error {
	s.UpdatedAt = nowISO()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal state: %w", err)
	}
	dir := filepath.Dir(StateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create %s: %w", dir, err)
	}
	if err := os.WriteFile(StateFile, data, 0644); err != nil {
		return fmt.Errorf("cannot write %s: %w", StateFile, err)
	}
	return nil
}

func (s *State) SetPhase(phase string) {
	s.CurrentPhase = phase
	s.UpdatedAt = nowISO()
}

func (s *State) IsApproved(gate string) bool {
	return s.Approvals[gate]
}

func (s *State) Approve(gate string) {
	if s.Approvals == nil {
		s.Approvals = make(ApprovalMap)
	}
	s.Approvals[gate] = true
	s.UpdatedAt = nowISO()
}

func (s *State) SetBlocked(reason string) {
	s.Status = StatusBlocked
	s.BlockReason = reason
}

func (s *State) SetReady() {
	s.Status = StatusReady
	s.BlockReason = ""
}

func (s *State) SetClosed() {
	s.Status = StatusClosed
	s.BlockReason = ""
}
