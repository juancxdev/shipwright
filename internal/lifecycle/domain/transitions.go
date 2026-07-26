package domain

func DefaultTransitions(reviewArtifacts []string) []Transition {
	return []Transition{
		{From: StateIntake, To: StateDiscovery, Via: "start", Trigger: "Nueva petición registrada", RequiredArtifacts: []string{".harness/artifacts/product/discovery.md"}},
		{From: StateDiscovery, To: StateProductContextReady, Via: "next", Trigger: "PO completa discovery", RequiredArtifacts: []string{".harness/artifacts/product/context.md", ".harness/artifacts/product/assumptions.md", ".harness/artifacts/product/open-questions.md"}},
		{From: StateProductContextReady, To: StateTechnicalScopeDraft, Via: "next", Trigger: "TL analiza contexto", RequiredArtifacts: []string{".harness/artifacts/architecture/technology-options.md"}},
		{From: StateTechnicalScopeDraft, To: StateScopeReview, Via: "next", Trigger: "PO prepara explicación de alcance", RequiredArtifacts: []string{".harness/artifacts/product/scope.md"}},
		{From: StateScopeReview, To: StateScopeApproved, Via: "approve:" + GateScope, Trigger: "Usuario aprueba alcance", RequiredApproval: GateScope},
		{From: StateScopeReview, To: StateDiscovery, Via: "request-change", Trigger: "Usuario pide cambios al alcance"},
		{From: StateScopeApproved, To: StateProjectPlanning, Via: "next", Trigger: "PM genera plan", RequiredArtifacts: []string{".harness/artifacts/project/project-charter.md", ".harness/artifacts/project/project-plan.md", ".harness/artifacts/project/risk-register.md"}},
		{From: StateProjectPlanning, To: StateUXDecision, Via: "next", Trigger: "Evaluar necesidad de UI", RequiredArtifacts: []string{".harness/artifacts/project/delivery-plan.md"}},
		{From: StateUXDecision, To: StateUXDesign, Via: "next", Trigger: "Requiere UI — iniciar diseño", RequiredArtifacts: []string{".harness/artifacts/design/ux-brief.md"}, Condition: ConditionRequiresUI},
		{From: StateUXDecision, To: StateTechnicalDesign, Via: "next", Trigger: "No requiere UI — skip a diseño técnico", Condition: ConditionNoUI},
		{From: StateUXDesign, To: StateUXApproval, Via: "next", Trigger: "Diseño listo para aprobación", RequiredArtifacts: []string{".harness/artifacts/design/prototype.md", ".harness/artifacts/design/user-flows.md", ".harness/artifacts/design/responsive-qa.md"}},
		{From: StateUXApproval, To: StateTechnicalDesign, Via: "approve:" + GateUXDesign, Trigger: "Usuario aprueba diseño UX", RequiredApproval: GateUXDesign},
		{From: StateUXApproval, To: StateUXDesign, Via: "request-change", Trigger: "Usuario rechaza diseño UX"},
		{From: StateTechnicalDesign, To: StateBacklogReady, Via: "next", Trigger: "TL crea arquitectura, contratos y backlog", RequiredArtifacts: []string{".harness/artifacts/architecture/system-architecture.md", ".harness/artifacts/contracts/openapi.yaml", ".harness/artifacts/backlog/epics.md", ".harness/artifacts/backlog/user-stories.md", ".harness/artifacts/backlog/frontend-tasks.md", ".harness/artifacts/backlog/backend-tasks.md", ".harness/artifacts/sdd/proposal.md", ".harness/artifacts/sdd/spec.md", ".harness/artifacts/sdd/tasks.md"}},
		{From: StateBacklogReady, To: StateImplementation, Via: "approve:" + GateTechnicalPlan, Trigger: "Gate técnico aprobado — comenzar implementación", RequiredApproval: GateTechnicalPlan},
		{From: StateImplementation, To: StateIntegration, Via: "next", Trigger: "FE/BE completan tareas", RequiredArtifacts: []string{".harness/artifacts/progress/frontend.md", ".harness/artifacts/progress/backend.md"}},
		{From: StateIntegration, To: StateQASecurityReview, Via: "next", Trigger: "Integración candidata lista", RequiredArtifacts: []string{".harness/artifacts/reports/contract-test-report.md", ".harness/artifacts/reports/review-checklist.md"}},
		{From: StateQASecurityReview, To: StateTechLeadReview, Via: "next", Trigger: "Pasa QA/security review", RequiredArtifacts: reviewArtifacts},
		{From: StateQASecurityReview, To: StateImplementation, Via: "request-change", Trigger: "Fallas críticas — volver a implementación"},
		{From: StateTechLeadReview, To: StateUserAcceptance, Via: "approve:" + GateTechLeadReview, Trigger: "TL aprueba — enviar a aceptación de usuario", RequiredApproval: GateTechLeadReview},
		{From: StateTechLeadReview, To: StateImplementation, Via: "request-change", Trigger: "TL rechaza — volver a implementación"},
		{From: StateUserAcceptance, To: StateClosed, Via: "approve:" + GateFinalAcceptance, Trigger: "Usuario acepta entrega final", RequiredApproval: GateFinalAcceptance, RequiredArtifacts: []string{".harness/artifacts/project/acceptance-report.md"}},
		{From: StateUserAcceptance, To: StateChangeRequest, Via: "request-change", Trigger: "Usuario pide cambios — abrir change request"},
		{From: StateChangeRequest, To: StateDiscovery, Via: "next", Trigger: "Cambio grande — nueva discovery parcial", RequiredArtifacts: []string{".harness/artifacts/project/change-management.md"}},
	}
}

func FindTransitions(transitions []Transition, from string, via string) []Transition {
	var result []Transition
	for _, transition := range transitions {
		if transition.From == from && transition.Via == via {
			result = append(result, transition)
		}
	}
	return result
}

func FindApprovalTransition(transitions []Transition, from string, gate string) *Transition {
	via := "approve:" + gate
	for _, transition := range transitions {
		if transition.From == from && transition.Via == via {
			copy := transition
			return &copy
		}
	}
	return nil
}

func FindChangeTransition(transitions []Transition, from string) *Transition {
	for _, transition := range transitions {
		if transition.From == from && transition.Via == "request-change" {
			copy := transition
			return &copy
		}
	}
	return nil
}

func GateForState(transitions []Transition, state string) string {
	for _, transition := range transitions {
		if transition.From == state && transition.RequiredApproval != "" {
			return transition.RequiredApproval
		}
	}
	return ""
}
