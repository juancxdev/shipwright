package harness

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const HandoffLogFile = "progress/handoffs.md"

type HandoffRecord struct {
	FromAgent string
	ToAgent   string
	Phase     string
	Artifact  string
	Action    string
	Reason    string
	Timestamp string
}

func LogHandoff(record HandoffRecord) error {
	if record.Timestamp == "" {
		record.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	entry := formatHandoffEntry(record)

	if err := ensureHandoffHeader(); err != nil {
		return err
	}

	f, err := os.OpenFile(HandoffLogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", HandoffLogFile, err)
	}
	defer f.Close()

	_, err = f.WriteString(entry)
	return err
}

func formatHandoffEntry(r HandoffRecord) string {
	var sb strings.Builder

	sb.WriteString("\n---\n\n")
	sb.WriteString(fmt.Sprintf("## Handoff — %s\n\n", r.Timestamp))
	sb.WriteString(fmt.Sprintf("- **From:** %s\n", r.FromAgent))
	sb.WriteString(fmt.Sprintf("- **To:** %s\n", r.ToAgent))
	sb.WriteString(fmt.Sprintf("- **Phase:** %s\n", r.Phase))
	if r.Artifact != "" {
		sb.WriteString(fmt.Sprintf("- **Artifact:** %s\n", r.Artifact))
	}
	sb.WriteString(fmt.Sprintf("- **Action:** %s\n", r.Action))
	if r.Reason != "" {
		sb.WriteString(fmt.Sprintf("- **Reason:** %s\n", r.Reason))
	}

	return sb.String()
}

func ensureHandoffHeader() error {
	info, err := os.Stat(HandoffLogFile)
	if err == nil && info.Size() > 0 {
		return nil
	}

	header := `# Handoff Log

This file records every agent handoff in the delivery cycle.
Every handoff goes to file — no phone-game allowed.

Format: timestamp, from-agent, to-agent, phase, artifact, action, reason.
`
	return WriteFile(HandoffLogFile, header)
}

type PermissionCheck struct {
	Agent    *Agent
	Artifact string
	Allowed  bool
	Reason   string
}

func CheckAgentPermission(agent *Agent, artifact string) PermissionCheck {
	if agent == nil {
		return PermissionCheck{
			Artifact: artifact,
			Allowed:  false,
			Reason:   "no active agent for this phase",
		}
	}

	if agent.CanModifyArtifact(artifact) {
		return PermissionCheck{
			Agent:    agent,
			Artifact: artifact,
			Allowed:  true,
			Reason:   fmt.Sprintf("%s can modify %s", agent.Name, artifact),
		}
	}

	if agent.CanReadArtifact(artifact) {
		return PermissionCheck{
			Agent:    agent,
			Artifact: artifact,
			Allowed:  false,
			Reason:   fmt.Sprintf("%s can read but NOT modify %s", agent.Name, artifact),
		}
	}

	return PermissionCheck{
		Agent:    agent,
		Artifact: artifact,
		Allowed:  false,
		Reason:   fmt.Sprintf("%s has no access to %s", agent.Name, artifact),
	}
}

func CountHandoffs() int {
	data, err := os.ReadFile(HandoffLogFile)
	if err != nil {
		return 0
	}
	return strings.Count(string(data), "\n---\n")
}
