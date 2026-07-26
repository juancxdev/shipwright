package harness

import (
	"fmt"
	"os"
)

type StateRecoveryResult struct {
	RecoveredState *State
	BackupFile     string
	Message        string
}

func RecoverCorruptState(projectName string) (*StateRecoveryResult, error) {
	backup := fmt.Sprintf("%s.corrupt.%s.bak", StateFile, NowISOCompact())
	if _, err := os.Stat(StateFile); err == nil {
		data, readErr := os.ReadFile(StateFile)
		if readErr != nil {
			return nil, fmt.Errorf("cannot read corrupt state for backup: %w", readErr)
		}
		if err := WriteFile(backup, string(data)); err != nil {
			return nil, fmt.Errorf("cannot backup corrupt state: %w", err)
		}
	}

	if projectName == "" {
		projectName = "Recovered Project"
	}
	state := NewState(projectName)
	state.SetBlocked("state.json was corrupt and has been recovered; inspect backup before continuing")
	if err := state.Save(); err != nil {
		return nil, fmt.Errorf("cannot save recovered state: %w", err)
	}
	_ = AppendTransitionAudit(TransitionAuditEvent{
		Action: "recover-state",
		From:   "corrupt",
		To:     state.CurrentPhase,
		Result: "recovered",
		Reason: "state.json parse failure or manual recovery",
	})
	return &StateRecoveryResult{
		RecoveredState: state,
		BackupFile:     backup,
		Message:        "Recovered state.json and backed up corrupt file",
	}, nil
}
