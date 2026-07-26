package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	CurrentFile = ".harness/artifacts/progress/current.md"
	HistoryFile = ".harness/artifacts/progress/history.md"
)

func InitProgress() error {
	if err := os.MkdirAll(filepath.Dir(CurrentFile), 0755); err != nil {
		return err
	}
	current := mustRenderTemplate("templates/project/harness/runtime/progress-current-initial.md", nil)
	if err := os.WriteFile(CurrentFile, []byte(current), 0644); err != nil {
		return err
	}

	history := mustRenderTemplate("templates/project/harness/runtime/progress-history-initial.md", RenderVars{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	return os.WriteFile(HistoryFile, []byte(history), 0644)
}

func AppendHistory(event, phase, details string) error {
	entry := fmt.Sprintf("| %s | %s | %s | %s |\n",
		time.Now().UTC().Format(time.RFC3339),
		escapePipe(event),
		escapePipe(phase),
		escapePipe(details),
	)

	f, err := os.OpenFile(HistoryFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(entry)
	return err
}

func UpdateCurrent(s *State, nextAction string) error {
	var projectLine string
	if s.ProjectName != "" {
		projectLine = fmt.Sprintf("**Project:** %s\n", s.ProjectName)
	}
	var requestLine string
	if s.InitialRequest != "" {
		requestLine = fmt.Sprintf("**Request:** %s\n", s.InitialRequest)
	}
	var blockReasonLine string
	if s.BlockReason != "" {
		blockReasonLine = fmt.Sprintf("**Block reason:** %s\n", s.BlockReason)
	}
	var approvalsSection strings.Builder
	gates := []struct {
		key   string
		label string
	}{
		{GateScope, "scope"},
		{GateUXDesign, "ux-design"},
		{GateTechnicalPlan, "technical-plan"},
		{GateTechLeadReview, "tech-lead"},
		{GateFinalAcceptance, "final-acceptance"},
	}
	for _, g := range gates {
		mark := "[ ]"
		if s.IsApproved(g.key) {
			mark = "[x]"
		}
		approvalsSection.WriteString(fmt.Sprintf("- %s %s\n", mark, g.label))
	}

	var activeCRSection string
	if s.ActiveChangeRequest != nil && *s.ActiveChangeRequest != "" {
		activeCRSection = fmt.Sprintf("\n**Active change request:** %s\n", *s.ActiveChangeRequest)
	}

	var requiresUISection string
	if s.RequiresUI != nil {
		ui := "no"
		if *s.RequiresUI {
			ui = "yes"
		}
		requiresUISection = fmt.Sprintf("\n**Requires UI:** %s\n", ui)
	}

	var nextActionSection string
	if nextAction != "" {
		nextActionSection = "\n## Next action\n\n" + nextAction + "\n"
	}

	content := mustRenderTemplate("templates/project/harness/runtime/progress-current.md", RenderVars{
		"phase":               s.CurrentPhase,
		"status":              s.Status,
		"project_line":        projectLine,
		"request_line":        requestLine,
		"block_reason_line":   blockReasonLine,
		"approvals_section":   approvalsSection.String(),
		"active_cr_section":   activeCRSection,
		"requires_ui_section": requiresUISection,
		"next_action_section": nextActionSection,
	})
	return os.WriteFile(CurrentFile, []byte(content), 0644)
}

func escapePipe(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}
