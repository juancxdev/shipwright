package domain

const (
	StateIntake              = "INTAKE"
	StateDiscovery           = "DISCOVERY"
	StateProductContextReady = "PRODUCT_CONTEXT_READY"
	StateTechnicalScopeDraft = "TECHNICAL_SCOPE_DRAFT"
	StateScopeReview         = "SCOPE_REVIEW"
	StateScopeApproved       = "SCOPE_APPROVED"
	StateProjectPlanning     = "PROJECT_PLANNING"
	StateUXDecision          = "UX_DECISION"
	StateUXDesign            = "UX_DESIGN"
	StateUXApproval          = "UX_APPROVAL"
	StateTechnicalDesign     = "TECHNICAL_DESIGN"
	StateBacklogReady        = "BACKLOG_READY"
	StateImplementation      = "IMPLEMENTATION"
	StateIntegration         = "INTEGRATION"
	StateQASecurityReview    = "QA_SECURITY_REVIEW"
	StateTechLeadReview      = "TECH_LEAD_REVIEW"
	StateUserAcceptance      = "USER_ACCEPTANCE"
	StateClosed              = "CLOSED"
	StateChangeRequest       = "CHANGE_REQUEST"
)

const (
	GateScope           = "scope"
	GateUXDesign        = "ux-design"
	GateTechnicalPlan   = "technical-plan"
	GateTechLeadReview  = "tech-lead"
	GateFinalAcceptance = "final-acceptance"
)

const (
	ConditionRequiresUI = "requires_ui"
	ConditionNoUI       = "no_ui"
	ConditionNone       = ""
)

type Phase string

type Gate string

type Transition struct {
	From              string
	To                string
	Via               string
	Trigger           string
	RequiredArtifacts []string
	RequiredApproval  string
	Condition         string
}

func AllStates() []string {
	return []string{
		StateIntake,
		StateDiscovery,
		StateProductContextReady,
		StateTechnicalScopeDraft,
		StateScopeReview,
		StateScopeApproved,
		StateProjectPlanning,
		StateUXDecision,
		StateUXDesign,
		StateUXApproval,
		StateTechnicalDesign,
		StateBacklogReady,
		StateImplementation,
		StateIntegration,
		StateQASecurityReview,
		StateTechLeadReview,
		StateUserAcceptance,
		StateClosed,
		StateChangeRequest,
	}
}

func AllGates() []string {
	return []string{GateScope, GateUXDesign, GateTechnicalPlan, GateTechLeadReview, GateFinalAcceptance}
}

func BlockingStates() map[string]bool {
	return map[string]bool{
		StateDiscovery:      true,
		StateScopeReview:    true,
		StateUXApproval:     true,
		StateUserAcceptance: true,
		StateChangeRequest:  true,
	}
}

func StateExists(state string) bool {
	for _, candidate := range AllStates() {
		if candidate == state {
			return true
		}
	}
	return false
}

func IsBlocking(state string) bool {
	return BlockingStates()[state]
}

func StateIndex(state string) int {
	for i, candidate := range AllStates() {
		if candidate == state {
			return i
		}
	}
	return -1
}
