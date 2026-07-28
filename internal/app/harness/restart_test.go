package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRestartRequestBacksUpStateAndStartsDiscovery(t *testing.T) {
	withTempWorkingDir(t)

	state := NewState("KD Council")
	state.ProjectID = "kdcouncil-app"
	state.InitialRequest = "old request"
	state.SetPhase(StateImplementation)
	state.Approve(GateScope)
	requiresUI := true
	state.RequiresUI = &requiresUI
	if err := state.Save(); err != nil {
		t.Fatalf("save state: %v", err)
	}
	if err := WriteFile(".harness/artifacts/product/discovery.md", "old discovery"); err != nil {
		t.Fatalf("write discovery: %v", err)
	}

	result, err := RestartRequest(state, "new onboarding request")
	if err != nil {
		t.Fatalf("RestartRequest: %v", err)
	}
	if result.PreviousFrom != StateImplementation {
		t.Fatalf("previous = %s", result.PreviousFrom)
	}
	if result.State.ProjectID != "kdcouncil-app" {
		t.Fatalf("project id = %s", result.State.ProjectID)
	}
	if result.State.CurrentPhase != StateDiscovery {
		t.Fatalf("phase = %s", result.State.CurrentPhase)
	}
	if result.State.InitialRequest != "new onboarding request" {
		t.Fatalf("request = %s", result.State.InitialRequest)
	}
	if result.State.IsApproved(GateScope) {
		t.Fatal("scope approval should reset")
	}
	if result.State.RequiresUI != nil {
		t.Fatal("requires_ui should reset")
	}

	backupState, err := os.ReadFile(filepath.Join(result.BackupDir, "state.json"))
	if err != nil {
		t.Fatalf("read backup state: %v", err)
	}
	if !strings.Contains(string(backupState), `"current_phase": "IMPLEMENTATION"`) {
		t.Fatalf("backup state missing previous phase:\n%s", backupState)
	}
	backupDiscovery, err := os.ReadFile(filepath.Join(result.BackupDir, "discovery.md"))
	if err != nil {
		t.Fatalf("read backup discovery: %v", err)
	}
	if string(backupDiscovery) != "old discovery" {
		t.Fatalf("backup discovery = %q", backupDiscovery)
	}
	newDiscovery, err := os.ReadFile(".harness/artifacts/product/discovery.md")
	if err != nil {
		t.Fatalf("read new discovery: %v", err)
	}
	if !strings.Contains(string(newDiscovery), "new onboarding request") {
		t.Fatalf("new discovery missing request:\n%s", newDiscovery)
	}
}
