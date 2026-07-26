package domain

import "testing"

func TestDefaultTransitionsFindsUXApprovalGate(t *testing.T) {
	transitions := DefaultTransitions([]string{"qa.md"})
	transition := FindApprovalTransition(transitions, StateUXApproval, GateUXDesign)
	if transition == nil {
		t.Fatal("expected UX approval transition")
	}
	if transition.To != StateTechnicalDesign {
		t.Fatalf("to = %s, want %s", transition.To, StateTechnicalDesign)
	}
}

func TestLifecycleBlockingStates(t *testing.T) {
	if !IsBlocking(StateUXApproval) {
		t.Fatal("UX approval must be blocking")
	}
	if IsBlocking(StateTechnicalDesign) {
		t.Fatal("technical design should not be blocking")
	}
}
