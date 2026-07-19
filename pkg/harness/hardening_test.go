package harness

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestRepairStateInitializesMissingFields(t *testing.T) {
	state := &State{ProjectName: "Billing", CurrentPhase: "BROKEN", Status: "weird"}

	result := RepairState(state)
	if !result.Valid {
		t.Fatalf("expected repaired state to be valid, errors: %v", result.Errors)
	}
	if state.ProjectID == "" {
		t.Fatal("expected project_id to be generated")
	}
	if state.CurrentPhase != StateIntake {
		t.Fatalf("phase = %s, want %s", state.CurrentPhase, StateIntake)
	}
	if state.Status != StatusBlocked {
		t.Fatalf("status = %s, want %s", state.Status, StatusBlocked)
	}
	if state.BlockReason == "" {
		t.Fatal("expected block reason after recovering invalid phase")
	}
	for _, gate := range AllGates() {
		if _, ok := state.Approvals[gate]; !ok {
			t.Fatalf("missing gate after repair: %s", gate)
		}
	}
}

func TestLoadStateRepairsSemanticCorruption(t *testing.T) {
	chdirTemp(t)
	writeTestFile(t, StateFile, `{
  "project_name": "Billing",
  "current_phase": "NOPE",
  "status": "strange"
}`)

	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState returned error: %v", err)
	}
	if state.CurrentPhase != StateIntake {
		t.Fatalf("phase = %s, want INTAKE", state.CurrentPhase)
	}
	if state.ProjectID == "" {
		t.Fatal("expected project_id recovery")
	}

	reloaded, err := os.ReadFile(StateFile)
	if err != nil {
		t.Fatalf("read repaired state: %v", err)
	}
	if !strings.Contains(string(reloaded), `"current_phase": "INTAKE"`) {
		t.Fatalf("expected repaired state to be saved, got:\n%s", string(reloaded))
	}
}

func TestRecoverCorruptStateBacksUpAndCreatesSafeState(t *testing.T) {
	chdirTemp(t)
	writeTestFile(t, StateFile, `{ definitely not json`)

	result, err := RecoverCorruptState("Recovered Billing")
	if err != nil {
		t.Fatalf("RecoverCorruptState error: %v", err)
	}
	if result.BackupFile == "" {
		t.Fatal("expected backup file")
	}
	if _, err := os.Stat(result.BackupFile); err != nil {
		t.Fatalf("expected backup to exist: %v", err)
	}
	if result.RecoveredState.ProjectName != "Recovered Billing" {
		t.Fatalf("project = %s", result.RecoveredState.ProjectName)
	}
	if result.RecoveredState.Status != StatusBlocked {
		t.Fatalf("status = %s, want blocked", result.RecoveredState.Status)
	}
}

func TestStateMachineRequiresArtifactsBeforeNext(t *testing.T) {
	chdirTemp(t)
	state := NewState("Billing")
	state.CurrentPhase = StateDiscovery

	result := Advance(state)
	if result.Transitioned {
		t.Fatal("expected transition to be blocked without discovery artifacts")
	}
	if len(result.MissingArtifacts) == 0 {
		t.Fatal("expected missing artifacts")
	}
}

func TestUXDesignRequiresResponsiveQABeforeApproval(t *testing.T) {
	chdirTemp(t)
	state := NewState("Billing")
	state.CurrentPhase = StateUXDesign
	writeTestFile(t, "design/prototype.md", "prototype")
	writeTestFile(t, "design/user-flows.md", "flows")

	result := Advance(state)
	if result.Transitioned {
		t.Fatal("expected UX_DESIGN transition to be blocked without responsive QA")
	}
	if !containsString(result.MissingArtifacts, "design/responsive-qa.md") {
		t.Fatalf("missing artifacts = %v, want design/responsive-qa.md", result.MissingArtifacts)
	}
}

func TestGateApprovalRequiresCorrectPhase(t *testing.T) {
	state := NewState("Billing")
	state.CurrentPhase = StateDiscovery

	result := ApproveGate(state, GateScope)
	if result.Error == "" {
		t.Fatal("expected approval to fail in wrong phase")
	}
}

func TestTransitionAuditWrittenOnSuccessfulAdvance(t *testing.T) {
	chdirTemp(t)
	state := NewState("Billing")
	state.CurrentPhase = StateDiscovery
	writeTestFile(t, "product/context.md", "context")
	writeTestFile(t, "product/assumptions.md", "assumptions")
	writeTestFile(t, "product/open-questions.md", "none")

	result := Advance(state)
	if !result.Transitioned {
		t.Fatalf("expected transition, got block: %s", result.BlockReason)
	}
	data, err := os.ReadFile(TransitionAuditFile)
	if err != nil {
		t.Fatalf("expected audit file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, `"action":"next"`) || !strings.Contains(content, `"result":"transitioned"`) {
		t.Fatalf("unexpected audit content: %s", content)
	}
}

type fakeMemoryPort struct {
	name  string
	err   error
	saved []*MemoryEvent
}

func (f *fakeMemoryPort) Save(event *MemoryEvent) error {
	if f.err != nil {
		return f.err
	}
	f.saved = append(f.saved, event)
	return nil
}

func (f *fakeMemoryPort) AdapterName() string { return f.name }

func TestMemoryServiceFallsBackWhenPrimaryFails(t *testing.T) {
	chdirTemp(t)
	primary := &fakeMemoryPort{name: EngramMode, err: errors.New("boom")}
	fallback := &fakeMemoryPort{name: FallbackMode}
	service := &MemoryService{primary: primary, fallback: fallback, engramOn: true}

	err := service.Save(&MemoryEvent{Title: "Decision", Type: MemTypeDecision, Content: "content"})
	if err != nil {
		t.Fatalf("Save error: %v", err)
	}
	if len(primary.saved) != 0 {
		t.Fatal("primary should fail before saving")
	}
	if len(fallback.saved) != 1 {
		t.Fatalf("fallback saves = %d, want 1", len(fallback.saved))
	}
	if fallback.saved[0].SavedVia != FallbackMode {
		t.Fatalf("saved_via = %s, want fallback", fallback.saved[0].SavedVia)
	}
}

type fakeDesignPort struct {
	name   string
	err    error
	result *DesignResult
	status *DesignStatus
}

func (f *fakeDesignPort) StartDesign(state *State, request string) (*DesignResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

func (f *fakeDesignPort) Status() (*DesignStatus, error) { return f.status, nil }
func (f *fakeDesignPort) AdapterName() string            { return f.name }

func TestDesignServiceFallsBackWhenOpenPencilFails(t *testing.T) {
	chdirTemp(t)
	state := NewState("Billing")
	primary := &fakeDesignPort{name: DesignModeOpenPencil, err: errors.New("canvas unavailable")}
	fallback := &fakeDesignPort{name: DesignModeDocOnly, result: &DesignResult{Adapter: DesignModeDocOnly, Mode: DesignModeDocOnly}}
	service := &DesignService{primary: primary, fallback: fallback, openpencilOn: true}

	result, err := service.StartDesign(state, "design billing")
	if err != nil {
		t.Fatalf("StartDesign error: %v", err)
	}
	if !result.FallbackUsed {
		t.Fatal("expected fallback to be used")
	}
	if !strings.Contains(result.Message, "OpenPencil unavailable") {
		t.Fatalf("unexpected fallback message: %s", result.Message)
	}
}

func TestDocOnlyDesignFallbackGeneratesResponsiveQA(t *testing.T) {
	chdirTemp(t)
	state := NewState("Billing")

	result, err := NewDocOnlyDesignFallback().StartDesign(state, "billing design")
	if err != nil {
		t.Fatalf("StartDesign: %v", err)
	}
	if !containsString(result.FilesCreated, "design/responsive-qa.md") {
		t.Fatalf("files = %v, want responsive QA artifact", result.FilesCreated)
	}
	assertFileContainsLocal(t, "design/responsive-qa.md", "Mobile 390x844")
	assertFileContainsLocal(t, "design/responsive-qa.md", "No component extends outside")
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func assertFileContainsLocal(t *testing.T, path string, needle string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(data), needle) {
		t.Fatalf("%s does not contain %q", path, needle)
	}
}
