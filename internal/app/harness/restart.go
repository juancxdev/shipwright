package harness

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type RestartResult struct {
	State        *State
	PreviousFrom string
	BackupDir    string
}

func RestartRequest(current *State, request string) (*RestartResult, error) {
	if current == nil {
		return nil, errors.New("state is required")
	}
	request = strings.TrimSpace(request)
	if request == "" {
		return nil, errors.New("la petición no puede estar vacía")
	}

	backupDir := filepath.Join(StateDir, "archive", "restarts", NowISOCompact())
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create restart backup dir: %w", err)
	}
	if err := copyFileIfExists(StateFile, filepath.Join(backupDir, "state.json")); err != nil {
		return nil, err
	}
	if err := copyFileIfExists(".harness/artifacts/product/discovery.md", filepath.Join(backupDir, "discovery.md")); err != nil {
		return nil, err
	}

	next := NewState(current.ProjectName)
	next.ProjectID = current.ProjectID
	if next.ProjectID == "" {
		next.ProjectID = generateProjectID(current.ProjectName)
	}
	if err := StartRequest(next, request); err != nil {
		return nil, err
	}
	if err := AppendTransitionAudit(TransitionAuditEvent{
		Action: "restart",
		From:   current.CurrentPhase,
		To:     next.CurrentPhase,
		Result: "transitioned",
		Reason: "delivery cycle restarted with new request",
	}); err != nil {
		return nil, err
	}

	return &RestartResult{
		State:        next,
		PreviousFrom: current.CurrentPhase,
		BackupDir:    backupDir,
	}, nil
}

func copyFileIfExists(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("cannot read %s: %w", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("cannot create %s: %w", filepath.Dir(dst), err)
	}
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("cannot write %s: %w", dst, err)
	}
	return nil
}
