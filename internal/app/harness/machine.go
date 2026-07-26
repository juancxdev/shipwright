package harness

import lifecycle "shipwright/internal/lifecycle/domain"

const (
	StateIntake              = lifecycle.StateIntake
	StateDiscovery           = lifecycle.StateDiscovery
	StateProductContextReady = lifecycle.StateProductContextReady
	StateTechnicalScopeDraft = lifecycle.StateTechnicalScopeDraft
	StateScopeReview         = lifecycle.StateScopeReview
	StateScopeApproved       = lifecycle.StateScopeApproved
	StateProjectPlanning     = lifecycle.StateProjectPlanning
	StateUXDecision          = lifecycle.StateUXDecision
	StateUXDesign            = lifecycle.StateUXDesign
	StateUXApproval          = lifecycle.StateUXApproval
	StateTechnicalDesign     = lifecycle.StateTechnicalDesign
	StateBacklogReady        = lifecycle.StateBacklogReady
	StateImplementation      = lifecycle.StateImplementation
	StateIntegration         = lifecycle.StateIntegration
	StateQASecurityReview    = lifecycle.StateQASecurityReview
	StateTechLeadReview      = lifecycle.StateTechLeadReview
	StateUserAcceptance      = lifecycle.StateUserAcceptance
	StateClosed              = lifecycle.StateClosed
	StateChangeRequest       = lifecycle.StateChangeRequest
)

const (
	GateScope           = lifecycle.GateScope
	GateUXDesign        = lifecycle.GateUXDesign
	GateTechnicalPlan   = lifecycle.GateTechnicalPlan
	GateTechLeadReview  = lifecycle.GateTechLeadReview
	GateFinalAcceptance = lifecycle.GateFinalAcceptance
)

const (
	ConditionRequiresUI = lifecycle.ConditionRequiresUI
	ConditionNoUI       = lifecycle.ConditionNoUI
	ConditionNone       = lifecycle.ConditionNone
)

type Transition = lifecycle.Transition

var AllStates = lifecycle.AllStates()
var BlockingStates = lifecycle.BlockingStates()

var transitions = lifecycle.DefaultTransitions(RequiredReviewArtifacts())

func FindTransitions(from string, via string) []Transition {
	return lifecycle.FindTransitions(transitions, from, via)
}

func FindApprovalTransition(from string, gate string) *Transition {
	return lifecycle.FindApprovalTransition(transitions, from, gate)
}

func FindChangeTransition(from string) *Transition {
	return lifecycle.FindChangeTransition(transitions, from)
}

func FindNextTransitions(from string) []Transition {
	return FindTransitions(from, "next")
}

func StateExists(state string) bool {
	return lifecycle.StateExists(state)
}

func IsBlocking(state string) bool {
	return lifecycle.IsBlocking(state)
}

func GateForState(state string) string {
	return lifecycle.GateForState(transitions, state)
}

func StateIndex(state string) int {
	return lifecycle.StateIndex(state)
}
