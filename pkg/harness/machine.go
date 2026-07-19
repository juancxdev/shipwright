package harness

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

type Transition struct {
	From              string
	To                string
	Via               string
	Trigger           string
	RequiredArtifacts []string
	RequiredApproval  string
	Condition         string
}

var AllStates = []string{
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

var BlockingStates = map[string]bool{
	StateDiscovery:      true,
	StateScopeReview:    true,
	StateUXApproval:     true,
	StateUserAcceptance: true,
	StateChangeRequest:  true,
}

var transitions = []Transition{
	{
		From: StateIntake, To: StateDiscovery, Via: "start",
		Trigger:           "Nueva petición registrada",
		RequiredArtifacts: []string{"product/discovery.md"},
	},
	{
		From: StateDiscovery, To: StateProductContextReady, Via: "next",
		Trigger:           "PO completa discovery",
		RequiredArtifacts: []string{"product/context.md", "product/assumptions.md", "product/open-questions.md"},
	},
	{
		From: StateProductContextReady, To: StateTechnicalScopeDraft, Via: "next",
		Trigger:           "TL analiza contexto",
		RequiredArtifacts: []string{"architecture/technology-options.md"},
	},
	{
		From: StateTechnicalScopeDraft, To: StateScopeReview, Via: "next",
		Trigger:           "PO prepara explicación de alcance",
		RequiredArtifacts: []string{"product/scope.md"},
	},
	{
		From: StateScopeReview, To: StateScopeApproved, Via: "approve:" + GateScope,
		Trigger:          "Usuario aprueba alcance",
		RequiredApproval: GateScope,
	},
	{
		From: StateScopeReview, To: StateDiscovery, Via: "request-change",
		Trigger: "Usuario pide cambios al alcance",
	},
	{
		From: StateScopeApproved, To: StateProjectPlanning, Via: "next",
		Trigger:           "PM genera plan",
		RequiredArtifacts: []string{"project/project-charter.md", "project/project-plan.md", "project/risk-register.md"},
	},
	{
		From: StateProjectPlanning, To: StateUXDecision, Via: "next",
		Trigger:           "Evaluar necesidad de UI",
		RequiredArtifacts: []string{"project/delivery-plan.md"},
	},
	{
		From: StateUXDecision, To: StateUXDesign, Via: "next",
		Trigger:           "Requiere UI — iniciar diseño",
		RequiredArtifacts: []string{"design/ux-brief.md"},
		Condition:         ConditionRequiresUI,
	},
	{
		From: StateUXDecision, To: StateTechnicalDesign, Via: "next",
		Trigger:   "No requiere UI — skip a diseño técnico",
		Condition: ConditionNoUI,
	},
	{
		From: StateUXDesign, To: StateUXApproval, Via: "next",
		Trigger:           "Diseño listo para aprobación",
		RequiredArtifacts: []string{"design/prototype.md", "design/user-flows.md", "design/responsive-qa.md"},
	},
	{
		From: StateUXApproval, To: StateTechnicalDesign, Via: "approve:" + GateUXDesign,
		Trigger:          "Usuario aprueba diseño UX",
		RequiredApproval: GateUXDesign,
	},
	{
		From: StateUXApproval, To: StateUXDesign, Via: "request-change",
		Trigger: "Usuario rechaza diseño UX",
	},
	{
		From: StateTechnicalDesign, To: StateBacklogReady, Via: "next",
		Trigger: "TL crea arquitectura, contratos y backlog",
		RequiredArtifacts: []string{
			"architecture/system-architecture.md",
			"contracts/openapi.yaml",
			"backlog/epics.md",
			"backlog/user-stories.md",
			"backlog/frontend-tasks.md",
			"backlog/backend-tasks.md",
			"sdd/proposal.md",
			"sdd/spec.md",
			"sdd/tasks.md",
		},
	},
	{
		From: StateBacklogReady, To: StateImplementation, Via: "approve:" + GateTechnicalPlan,
		Trigger:          "Gate técnico aprobado — comenzar implementación",
		RequiredApproval: GateTechnicalPlan,
	},
	{
		From: StateImplementation, To: StateIntegration, Via: "next",
		Trigger:           "FE/BE completan tareas",
		RequiredArtifacts: []string{"progress/frontend.md", "progress/backend.md"},
	},
	{
		From: StateIntegration, To: StateQASecurityReview, Via: "next",
		Trigger:           "Integración candidata lista",
		RequiredArtifacts: []string{"reports/contract-test-report.md", "reports/review-checklist.md"},
	},
	{
		From: StateQASecurityReview, To: StateTechLeadReview, Via: "next",
		Trigger:           "Pasa QA/security review",
		RequiredArtifacts: RequiredReviewArtifacts(),
	},
	{
		From: StateQASecurityReview, To: StateImplementation, Via: "request-change",
		Trigger: "Fallas críticas — volver a implementación",
	},
	{
		From: StateTechLeadReview, To: StateUserAcceptance, Via: "approve:" + GateTechLeadReview,
		Trigger:          "TL aprueba — enviar a aceptación de usuario",
		RequiredApproval: GateTechLeadReview,
	},
	{
		From: StateTechLeadReview, To: StateImplementation, Via: "request-change",
		Trigger: "TL rechaza — volver a implementación",
	},
	{
		From: StateUserAcceptance, To: StateClosed, Via: "approve:" + GateFinalAcceptance,
		Trigger:           "Usuario acepta entrega final",
		RequiredApproval:  GateFinalAcceptance,
		RequiredArtifacts: []string{"project/acceptance-report.md"},
	},
	{
		From: StateUserAcceptance, To: StateChangeRequest, Via: "request-change",
		Trigger: "Usuario pide cambios — abrir change request",
	},
	{
		From: StateChangeRequest, To: StateDiscovery, Via: "next",
		Trigger:           "Cambio grande — nueva discovery parcial",
		RequiredArtifacts: []string{"project/change-management.md"},
	},
}

func FindTransitions(from string, via string) []Transition {
	var result []Transition
	for _, t := range transitions {
		if t.From == from && t.Via == via {
			result = append(result, t)
		}
	}
	return result
}

func FindApprovalTransition(from string, gate string) *Transition {
	via := "approve:" + gate
	for _, t := range transitions {
		if t.From == from && t.Via == via {
			return &t
		}
	}
	return nil
}

func FindChangeTransition(from string) *Transition {
	for _, t := range transitions {
		if t.From == from && t.Via == "request-change" {
			return &t
		}
	}
	return nil
}

func FindNextTransitions(from string) []Transition {
	return FindTransitions(from, "next")
}

func StateExists(state string) bool {
	for _, s := range AllStates {
		if s == state {
			return true
		}
	}
	return false
}

func IsBlocking(state string) bool {
	return BlockingStates[state]
}

func GateForState(state string) string {
	for _, t := range transitions {
		if t.From == state && t.RequiredApproval != "" {
			return t.RequiredApproval
		}
	}
	return ""
}

func StateIndex(state string) int {
	for i, s := range AllStates {
		if s == state {
			return i
		}
	}
	return -1
}
