package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ExecutorStateFile = ".harness/executor.json"

const (
	ExecutorGeneric  = "generic"
	ExecutorOpenCode = "opencode"
)

type ExecutorAdapter interface {
	Name() string
	Description() string
	Generate() (*ExecutorGenerateResult, error)
	Status() (*ExecutorStatus, error)
}

type ExecutorGenerateResult struct {
	Name         string   `json:"name"`
	FilesCreated []string `json:"files_created"`
	FilesUpdated []string `json:"files_updated"`
	Message      string   `json:"message"`
}

type ExecutorStatus struct {
	Name       string   `json:"name"`
	Configured bool     `json:"configured"`
	Files      []string `json:"files"`
	Missing    []string `json:"missing"`
	Warnings   []string `json:"warnings,omitempty"`
}

type ExecutorState struct {
	Executor        string `json:"executor"`
	LastGeneratedAt string `json:"last_generated_at"`
}

func ExecutorNames() []string {
	return []string{ExecutorGeneric, ExecutorOpenCode}
}

func GetExecutorAdapter(name string) (ExecutorAdapter, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", ExecutorGeneric:
		return GenericExecutorAdapter{}, nil
	case ExecutorOpenCode:
		return OpenCodeExecutorAdapter{}, nil
	default:
		return nil, fmt.Errorf("unknown executor %q (valid: generic | opencode)", name)
	}
}

func GenerateExecutor(name string) (*ExecutorGenerateResult, error) {
	adapter, err := GetExecutorAdapter(name)
	if err != nil {
		return nil, err
	}
	result, err := adapter.Generate()
	if err != nil {
		return nil, err
	}
	state := ExecutorState{Executor: adapter.Name(), LastGeneratedAt: NowISO()}
	if err := SaveExecutorState(state); err != nil {
		return nil, err
	}
	return result, nil
}

func SaveExecutorState(state ExecutorState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return WriteFile(ExecutorStateFile, string(data))
}

func LoadExecutorState() (*ExecutorState, error) {
	data, err := os.ReadFile(ExecutorStateFile)
	if err != nil {
		return nil, err
	}
	var state ExecutorState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func writeTrackedFile(path, content string, result *ExecutorGenerateResult) error {
	existed := ArtifactExists(path)
	if err := WriteFile(path, content); err != nil {
		return err
	}
	if existed {
		result.FilesUpdated = append(result.FilesUpdated, path)
	} else {
		result.FilesCreated = append(result.FilesCreated, path)
	}
	return nil
}

func writeExecutableTrackedFile(path, content string, result *ExecutorGenerateResult) error {
	if err := writeTrackedFile(path, content, result); err != nil {
		return err
	}
	return os.Chmod(path, 0755)
}

func requiredStatus(name string, files []string) *ExecutorStatus {
	status := &ExecutorStatus{Name: name, Files: files}
	for _, file := range files {
		if !ArtifactExists(file) {
			status.Missing = append(status.Missing, file)
		}
	}
	status.Configured = len(status.Missing) == 0
	return status
}

func opencodeAgentPath(name string) string {
	return filepath.Join(".opencode", "agents", name+".md")
}

func opencodeSkillPath(name string) string {
	return filepath.Join(".opencode", "skills", name, "SKILL.md")
}

func openCodeConfigPath() string {
	return filepath.Join(".opencode", "opencode.json")
}
