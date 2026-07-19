package harness

import (
	"strings"
	"testing"
)

func TestBuildTDDPolicyStrictFromProjectProfile(t *testing.T) {
	profile := &ProjectProfile{
		TDD:      TDDCapability{Supported: true, RecommendedMode: TDDModeStrict},
		Commands: ProjectCommands{Test: []DetectedCommand{{Command: "go test ./...", Source: "go.mod", Confidence: "high"}}},
	}

	policy := BuildTDDPolicy(profile)
	if policy.Mode != TDDModeStrict {
		t.Fatalf("mode = %s, want strict", policy.Mode)
	}
	if !policy.Supported {
		t.Fatal("policy should be supported")
	}
	if policy.TestCommand != "go test ./..." {
		t.Fatalf("test command = %q", policy.TestCommand)
	}
	if len(policy.RequiredEvidence) == 0 {
		t.Fatal("strict policy should require evidence")
	}
}

func TestTDDBlockReasonBlocksStrictWithoutEvidence(t *testing.T) {
	chdirTemp(t)
	policy := BuildTDDPolicy(&ProjectProfile{
		TDD:      TDDCapability{Supported: true, RecommendedMode: TDDModeStrict},
		Commands: ProjectCommands{Test: []DetectedCommand{{Command: "go test ./...", Source: "go.mod", Confidence: "high"}}},
	})
	if err := SaveTDDPolicy(policy); err != nil {
		t.Fatalf("SaveTDDPolicy: %v", err)
	}
	writeTestFile(t, "progress/frontend.md", "# Frontend progress\n\nImplemented screens.\n")
	writeTestFile(t, "progress/backend.md", "# Backend progress\n\nImplemented API.\n")

	reason := TDDBlockReason()
	if reason == "" {
		t.Fatal("expected strict TDD to block without evidence")
	}
	if !strings.Contains(reason, "Strict TDD gate blocked") {
		t.Fatalf("unexpected block reason:\n%s", reason)
	}
}

func TestTDDBlockReasonAllowsStrictWithProgressEvidence(t *testing.T) {
	chdirTemp(t)
	policy := BuildTDDPolicy(&ProjectProfile{
		TDD:      TDDCapability{Supported: true, RecommendedMode: TDDModeStrict},
		Commands: ProjectCommands{Test: []DetectedCommand{{Command: "go test ./...", Source: "go.mod", Confidence: "high"}}},
	})
	if err := SaveTDDPolicy(policy); err != nil {
		t.Fatalf("SaveTDDPolicy: %v", err)
	}
	writeTestFile(t, "progress/frontend.md", "# Frontend progress\n\n## TDD evidence:\nCommand: go test ./... PASS.\n")
	writeTestFile(t, "progress/backend.md", "# Backend progress\n\n## Test evidence:\nCommand: go test ./... PASS.\n")

	if reason := TDDBlockReason(); reason != "" {
		t.Fatalf("expected strict TDD to pass with evidence, got:\n%s", reason)
	}
}

func TestAdvanceBlocksImplementationWithoutStrictTDDEvidence(t *testing.T) {
	chdirTemp(t)
	state := NewState("tdd-project")
	state.SetPhase(StateImplementation)
	state.SetReady()
	if err := state.Save(); err != nil {
		t.Fatalf("state save: %v", err)
	}
	policy := BuildTDDPolicy(&ProjectProfile{
		TDD:      TDDCapability{Supported: true, RecommendedMode: TDDModeStrict},
		Commands: ProjectCommands{Test: []DetectedCommand{{Command: "go test ./...", Source: "go.mod", Confidence: "high"}}},
	})
	if err := SaveTDDPolicy(policy); err != nil {
		t.Fatalf("SaveTDDPolicy: %v", err)
	}
	writeTestFile(t, "progress/frontend.md", "# Frontend progress\n\nImplemented UI.\n")
	writeTestFile(t, "progress/backend.md", "# Backend progress\n\nImplemented API.\n")

	result := Advance(state)
	if result.Transitioned {
		t.Fatal("expected implementation transition to be blocked by strict TDD")
	}
	if result.To != StateIntegration {
		t.Fatalf("target = %s, want %s", result.To, StateIntegration)
	}
	if !strings.Contains(result.BlockReason, "Strict TDD gate blocked") {
		t.Fatalf("unexpected block reason:\n%s", result.BlockReason)
	}
}
