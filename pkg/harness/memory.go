package harness

import (
	"fmt"
	"strings"
	"time"
)

const (
	MemTypeDecision       = "decision"
	MemTypeDiscovery      = "discovery"
	MemTypeBugfix         = "bugfix"
	MemTypePattern        = "pattern"
	MemTypeArchitecture   = "architecture"
	MemTypeSessionSummary = "session_summary"
)

const (
	EngramMode   = "engram"
	FallbackMode = "fallback"
)

const MemoryQueueFile = ".harness/memory-queue.json"
const DecisionsFile = "progress/decisions.md"

type MemoryEvent struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	TopicKey  string `json:"topic_key"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	SavedVia  string `json:"saved_via"`
}

type MemoryPort interface {
	Save(event *MemoryEvent) error
	AdapterName() string
}

type MemoryService struct {
	primary  MemoryPort
	fallback MemoryPort
	engramOn bool
}

func NewMemoryService(integrations *Integrations) *MemoryService {
	engramOn := integrations != nil && integrations.Engram.Enabled

	svc := &MemoryService{
		fallback: NewLocalMemoryFallback(),
		engramOn: engramOn,
	}

	if engramOn {
		svc.primary = NewEngramMemoryAdapter()
	} else {
		svc.primary = svc.fallback
	}

	return svc
}

func (ms *MemoryService) Save(event *MemoryEvent) error {
	event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	if event.ID == "" {
		event.ID = generateMemoryID()
	}

	if ms.engramOn {
		err := ms.primary.Save(event)
		if err != nil {
			event.SavedVia = FallbackMode
			return ms.fallback.Save(event)
		}
		event.SavedVia = EngramMode
		return nil
	}

	event.SavedVia = FallbackMode
	return ms.primary.Save(event)
}

func (ms *MemoryService) SaveDecision(title, topicKey, what, why, where, learned string) error {
	return ms.Save(&MemoryEvent{
		Type:     MemTypeDecision,
		Title:    title,
		TopicKey: topicKey,
		Content:  formatMemoryContent(what, why, where, learned),
	})
}

func (ms *MemoryService) SaveDiscovery(title, topicKey, what, why, where, learned string) error {
	return ms.Save(&MemoryEvent{
		Type:     MemTypeDiscovery,
		Title:    title,
		TopicKey: topicKey,
		Content:  formatMemoryContent(what, why, where, learned),
	})
}

func (ms *MemoryService) SaveArchitecture(title, topicKey, what, why, where, learned string) error {
	return ms.Save(&MemoryEvent{
		Type:     MemTypeArchitecture,
		Title:    title,
		TopicKey: topicKey,
		Content:  formatMemoryContent(what, why, where, learned),
	})
}

func (ms *MemoryService) SavePattern(title, topicKey, what, why, where, learned string) error {
	return ms.Save(&MemoryEvent{
		Type:     MemTypePattern,
		Title:    title,
		TopicKey: topicKey,
		Content:  formatMemoryContent(what, why, where, learned),
	})
}

func (ms *MemoryService) SaveSessionSummary(title, content string) error {
	return ms.Save(&MemoryEvent{
		Type:     MemTypeSessionSummary,
		Title:    title,
		TopicKey: "session",
		Content:  content,
	})
}

func (ms *MemoryService) AdapterName() string {
	if ms.engramOn {
		return EngramMode
	}
	return FallbackMode
}

func (ms *MemoryService) IsEngramEnabled() bool {
	return ms.engramOn
}

func SaveGateMemory(ms *MemoryService, state *State, gate string) error {
	switch gate {
	case GateScope:
		return ms.SaveDecision(
			"Scope approved: "+state.ProjectName,
			"project/scope",
			"User approved the functional scope for the project",
			"Gate approval — scope review completed",
			"product/scope.md, .harness/approvals/scope.json",
			"",
		)

	case GateUXDesign:
		return ms.SaveDecision(
			"UX design approved: "+state.ProjectName,
			"design/ux-approval",
			"User approved the UX design for the project",
			"Gate approval — UX design review completed",
			"design/prototype.md, design/user-flows.md, .harness/approvals/ux-design.json",
			"",
		)

	case GateTechnicalPlan:
		return ms.SaveArchitecture(
			"Technical plan approved: "+state.ProjectName,
			"architecture/system",
			"User approved the technical plan including architecture and backlog",
			"Gate approval — technical plan review completed, implementation can begin",
			"architecture/system-architecture.md, backlog/epics.md, .harness/approvals/technical-plan.json",
			"",
		)

	case GateTechLeadReview:
		return ms.SaveDecision(
			"Tech lead review approved: "+state.ProjectName,
			"project/tech-lead-review",
			"Tech lead approved the implementation after QA/security review",
			"Gate approval — tech lead review passed, ready for user acceptance",
			"reports/qa-report.md, reports/security-review.md, .harness/approvals/tech-lead.json",
			"",
		)

	case GateFinalAcceptance:
		return ms.SaveDecision(
			"Final acceptance: "+state.ProjectName,
			"project/acceptance",
			"User accepted the final delivery of the project",
			"Gate approval — project closed with user acceptance",
			"project/acceptance-report.md, .harness/approvals/final-acceptance.json",
			"",
		)

	default:
		return nil
	}
}

func SaveChangeRequestMemory(ms *MemoryService, state *State, reason, crFile string) error {
	return ms.SaveDiscovery(
		"Change request: "+state.ProjectName,
		"project/change-request",
		"Change request created: "+reason,
		"User requested a change during the delivery cycle",
		crFile,
		"Change requests trigger re-evaluation of scope and impact",
	)
}

func SavePhaseTransitionMemory(ms *MemoryService, state *State, fromPhase, toPhase string) error {
	if toPhase == StateTechnicalDesign {
		return ms.SaveArchitecture(
			"Architecture phase entered: "+state.ProjectName,
			"architecture/system",
			"Project entered TECHNICAL_DESIGN phase — TL must create architecture, contracts and backlog",
			"State machine transition from "+fromPhase+" to "+toPhase,
			"architecture/, contracts/, backlog/",
			"",
		)
	}

	if toPhase == StateClosed {
		return ms.SaveSessionSummary(
			"Project closed: "+state.ProjectName,
			fmt.Sprintf("**What**: Project %s completed and accepted by user\n**Why**: Final acceptance gate approved\n**Where**: .harness/state.json", state.ProjectName),
		)
	}

	return nil
}

func formatMemoryContent(what, why, where, learned string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**What**: %s\n", what))
	sb.WriteString(fmt.Sprintf("**Why**: %s\n", why))
	sb.WriteString(fmt.Sprintf("**Where**: %s\n", where))
	if learned != "" {
		sb.WriteString(fmt.Sprintf("**Learned**: %s\n", learned))
	}
	return sb.String()
}

func generateMemoryID() string {
	return fmt.Sprintf("mem-%s", time.Now().UTC().Format("20060102-150405.000"))
}
