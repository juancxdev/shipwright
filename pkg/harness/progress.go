package harness

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	CurrentFile = "progress/current.md"
	HistoryFile = "progress/history.md"
)

func InitProgress() error {
	if err := os.MkdirAll("progress", 0755); err != nil {
		return err
	}
	current := fmt.Sprintf(`# Current Status

**Phase:** INTAKE
**Status:** ready
**Project:** (not set)

## Next action

Run ` + "`shipwright start \"<your request>\"`" + ` to begin discovery.
`)
	if err := os.WriteFile(CurrentFile, []byte(current), 0644); err != nil {
		return err
	}

	history := fmt.Sprintf(`# History

| Timestamp | Event | Phase | Details |
|---|---|---|---|
| %s | init | INTAKE | Harness initialized |
`, time.Now().UTC().Format(time.RFC3339))
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
	var sb strings.Builder

	sb.WriteString("# Current Status\n\n")
	sb.WriteString(fmt.Sprintf("**Phase:** %s\n", s.CurrentPhase))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", s.Status))
	if s.ProjectName != "" {
		sb.WriteString(fmt.Sprintf("**Project:** %s\n", s.ProjectName))
	}
	if s.InitialRequest != "" {
		sb.WriteString(fmt.Sprintf("**Request:** %s\n", s.InitialRequest))
	}
	if s.BlockReason != "" {
		sb.WriteString(fmt.Sprintf("**Block reason:** %s\n", s.BlockReason))
	}
	sb.WriteString("\n## Approvals\n\n")
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
		sb.WriteString(fmt.Sprintf("- %s %s\n", mark, g.label))
	}

	if s.ActiveChangeRequest != nil && *s.ActiveChangeRequest != "" {
		sb.WriteString(fmt.Sprintf("\n**Active change request:** %s\n", *s.ActiveChangeRequest))
	}

	if s.RequiresUI != nil {
		ui := "no"
		if *s.RequiresUI {
			ui = "yes"
		}
		sb.WriteString(fmt.Sprintf("\n**Requires UI:** %s\n", ui))
	}

	if nextAction != "" {
		sb.WriteString("\n## Next action\n\n")
		sb.WriteString(nextAction)
		sb.WriteString("\n")
	}

	return os.WriteFile(CurrentFile, []byte(sb.String()), 0644)
}

func escapePipe(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}
