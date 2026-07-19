package harness

import (
	"fmt"
	"strings"
	"time"
)

type AdvanceResult struct {
	Transitioned     bool
	From             string
	To               string
	Message          string
	MissingArtifacts []string
	BlockReason      string
}

func Advance(s *State) *AdvanceResult {
	nextTransitions := FindNextTransitions(s.CurrentPhase)

	if len(nextTransitions) == 0 {
		return &AdvanceResult{
			Transitioned: false,
			From:         s.CurrentPhase,
			Message:      fmt.Sprintf("No hay transición 'next' desde %s. Usá 'shipwright approve <gate>' o 'shipwright request-change' según corresponda.", s.CurrentPhase),
		}
	}

	var candidate *Transition
	for i := range nextTransitions {
		t := &nextTransitions[i]
		if t.Condition == ConditionNone {
			candidate = t
			break
		}
		if t.Condition == ConditionRequiresUI {
			if s.RequiresUI == nil {
				return &AdvanceResult{
					Transitioned: false,
					From:         s.CurrentPhase,
					BlockReason:  "No se decidió si el proyecto requiere UI. Editá .harness/state.json y seteá requires_ui a true o false.",
				}
			}
			if *s.RequiresUI {
				candidate = t
				break
			}
		}
		if t.Condition == ConditionNoUI {
			if s.RequiresUI == nil {
				return &AdvanceResult{
					Transitioned: false,
					From:         s.CurrentPhase,
					BlockReason:  "No se decidió si el proyecto requiere UI. Editá .harness/state.json y seteá requires_ui a true o false.",
				}
			}
			if !*s.RequiresUI {
				candidate = t
				break
			}
		}
	}

	if candidate == nil {
		return &AdvanceResult{
			Transitioned: false,
			From:         s.CurrentPhase,
			Message:      fmt.Sprintf("No se encontró una transición válida desde %s con la configuración actual (requires_ui=%v).", s.CurrentPhase, s.RequiresUI),
		}
	}

	missing := CheckArtifacts(candidate.RequiredArtifacts)
	if len(missing) > 0 {
		_ = AppendTransitionAudit(TransitionAuditEvent{
			Action: "next",
			From:   s.CurrentPhase,
			To:     candidate.To,
			Result: "blocked",
			Reason: fmt.Sprintf("missing artifacts: %s", strings.Join(missing, ", ")),
		})
		return &AdvanceResult{
			Transitioned:     false,
			From:             s.CurrentPhase,
			To:               candidate.To,
			MissingArtifacts: missing,
			BlockReason:      fmt.Sprintf("Faltan artefactos para avanzar a %s:\n  %s", candidate.To, formatFileList(missing)),
		}
	}

	if candidate.From == StateImplementation {
		if reason := TDDBlockReason(); reason != "" {
			_ = AppendTransitionAudit(TransitionAuditEvent{
				Action: "next",
				From:   s.CurrentPhase,
				To:     candidate.To,
				Result: "blocked",
				Reason: "strict TDD evidence blocked progress",
			})
			return &AdvanceResult{
				Transitioned: false,
				From:         s.CurrentPhase,
				To:           candidate.To,
				BlockReason:  reason,
			}
		}
	}

	if candidate.From == StateIntegration {
		if reason := ContractReviewBlockReason(); reason != "" {
			_ = AppendTransitionAudit(TransitionAuditEvent{
				Action: "next",
				From:   s.CurrentPhase,
				To:     candidate.To,
				Result: "blocked",
				Reason: "contract review evidence blocked progress",
			})
			return &AdvanceResult{
				Transitioned: false,
				From:         s.CurrentPhase,
				To:           candidate.To,
				BlockReason:  reason,
			}
		}
	}

	if candidate.From == StateQASecurityReview {
		if reason := ReviewBlockReason(); reason != "" {
			_ = AppendTransitionAudit(TransitionAuditEvent{
				Action: "next",
				From:   s.CurrentPhase,
				To:     candidate.To,
				Result: "blocked",
				Reason: "QA/security review evidence blocked progress",
			})
			return &AdvanceResult{
				Transitioned: false,
				From:         s.CurrentPhase,
				To:           candidate.To,
				BlockReason:  reason,
			}
		}
	}

	fromPhase := s.CurrentPhase
	s.SetPhase(candidate.To)
	if candidate.To == StateClosed {
		s.SetClosed()
	} else if IsBlocking(candidate.To) {
		s.SetBlocked(fmt.Sprintf("Estado %s requiere intervención humana.", candidate.To))
	} else {
		s.SetReady()
	}

	_ = AppendTransitionAudit(TransitionAuditEvent{
		Action: "next",
		From:   fromPhase,
		To:     candidate.To,
		Result: "transitioned",
		Reason: candidate.Trigger,
	})

	return &AdvanceResult{
		Transitioned: true,
		From:         fromPhase,
		To:           candidate.To,
		Message:      candidate.Trigger,
	}
}

type ApproveResult struct {
	Transitioned     bool
	From             string
	To               string
	Gate             string
	Message          string
	MissingArtifacts []string
	Error            string
}

func ApproveGate(s *State, gate string) *ApproveResult {
	t := FindApprovalTransition(s.CurrentPhase, gate)
	if t == nil {
		validState := findStateForGate(gate)
		return &ApproveResult{
			Error: fmt.Sprintf("No se puede aprobar '%s' en la fase %s. Este gate es válido en %s.", gate, s.CurrentPhase, validState),
		}
	}

	missing := CheckArtifacts(t.RequiredArtifacts)
	if len(missing) > 0 {
		_ = AppendTransitionAudit(TransitionAuditEvent{
			Action: "approve",
			From:   s.CurrentPhase,
			To:     t.To,
			Gate:   gate,
			Result: "blocked",
			Reason: fmt.Sprintf("missing artifacts: %s", strings.Join(missing, ", ")),
		})
		return &ApproveResult{
			Transitioned:     false,
			From:             s.CurrentPhase,
			To:               t.To,
			Gate:             gate,
			MissingArtifacts: missing,
			Error:            fmt.Sprintf("Faltan artefactos para aprobar %s:\n  %s", gate, formatFileList(missing)),
		}
	}

	fromPhase := s.CurrentPhase
	s.Approve(gate)
	s.SetPhase(t.To)

	if t.To == StateClosed {
		s.SetClosed()
	} else if IsBlocking(t.To) {
		s.SetBlocked(fmt.Sprintf("Estado %s requiere intervención humana.", t.To))
	} else {
		s.SetReady()
	}

	_ = AppendTransitionAudit(TransitionAuditEvent{
		Action: "approve",
		From:   fromPhase,
		To:     t.To,
		Gate:   gate,
		Result: "transitioned",
		Reason: t.Trigger,
	})

	return &ApproveResult{
		Transitioned: true,
		From:         fromPhase,
		To:           t.To,
		Gate:         gate,
		Message:      t.Trigger,
	}
}

type ChangeResult struct {
	Transitioned bool
	From         string
	To           string
	Message      string
	CRFile       string
	Error        string
}

func RequestChange(s *State, reason string) *ChangeResult {
	t := FindChangeTransition(s.CurrentPhase)
	if t == nil {
		return &ChangeResult{
			Error: fmt.Sprintf("No se puede pedir cambio desde %s.", s.CurrentPhase),
		}
	}

	crID := generateCRID()
	crFile := fmt.Sprintf("project/change-requests/CR-%s.md", crID)

	crContent := fmt.Sprintf(`# CR-%s — Change Request

## Solicitud

%s

## Motivo

(pendiente de completar)

## Impacto funcional

(pendiente)

## Impacto técnico

(pendiente)

## Impacto en alcance

(pendiente)

## Impacto en tiempo/esfuerzo

(pendiente)

## Riesgos

(pendiente)

## Decisión

- [ ] aprobado
- [ ] rechazado
- [ ] postergado

## Aprobado por

(pendiente)
`, crID, escapeForMarkdown(reason))

	if err := WriteFile(crFile, crContent); err != nil {
		return &ChangeResult{
			Error: fmt.Sprintf("No se pudo crear %s: %s", crFile, err),
		}
	}

	fromPhase := s.CurrentPhase
	s.SetPhase(t.To)
	crRef := fmt.Sprintf("CR-%s", crID)
	s.ActiveChangeRequest = &crRef
	s.SetBlocked(fmt.Sprintf("Change request %s creado. Transición a %s.", crRef, t.To))

	_ = AppendTransitionAudit(TransitionAuditEvent{
		Action: "request-change",
		From:   fromPhase,
		To:     t.To,
		Result: "transitioned",
		Reason: reason,
	})

	return &ChangeResult{
		Transitioned: true,
		From:         fromPhase,
		To:           t.To,
		Message:      t.Trigger,
		CRFile:       crFile,
	}
}

func StartRequest(s *State, request string) error {
	if s.CurrentPhase != StateIntake {
		return fmt.Errorf("el harness ya fue iniciado (fase actual: %s). Usá 'shipwright next' para avanzar.", s.CurrentPhase)
	}

	s.InitialRequest = request

	discoveryContent := fmt.Sprintf(`# Discovery

## Solicitud del usuario

%s

## Preguntas de discovery

(El Product Owner agent debe completar esta sección con preguntas para el usuario.

El harness NO avanza de DISCOVERY sin:

- product/context.md
- product/assumptions.md
- product/open-questions.md (sin preguntas críticas pendientes)

Completá esos archivos y ejecutá: shipwright next)

## Respuestas del usuario

(pendiente)
`, escapeForMarkdown(request))

	if err := WriteFile("product/discovery.md", discoveryContent); err != nil {
		return fmt.Errorf("no se pudo crear product/discovery.md: %w", err)
	}

	s.SetPhase(StateDiscovery)
	s.SetBlocked("Discovery iniciado. Product Owner debe preguntar al usuario en chat, registrar open questions y generar contexto/supuestos/scope antes de avanzar.")

	_ = AppendTransitionAudit(TransitionAuditEvent{
		Action: "start",
		From:   StateIntake,
		To:     StateDiscovery,
		Result: "transitioned",
		Reason: "initial request registered",
	})

	return nil
}

func findStateForGate(gate string) string {
	via := "approve:" + gate
	for _, t := range transitions {
		if t.Via == via {
			return t.From
		}
	}
	return "(desconocido)"
}

func generateCRID() string {
	return time.Now().UTC().Format("20060102-150405")
}

func formatFileList(files []string) string {
	return strings.Join(files, "\n  ")
}

func escapeForMarkdown(s string) string {
	s = strings.ReplaceAll(s, "`", "\\`")
	return s
}
