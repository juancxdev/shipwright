package application

import (
	"os"
	"path/filepath"
	"testing"
)

func chdirTemp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	return dir
}

func withTempWorkingDir(t *testing.T) string {
	t.Helper()
	return chdirTemp(t)
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func writeProjectProfileTestFile(t *testing.T, path, content string) {
	t.Helper()
	writeTestFile(t, path, content)
}

func CreateBaseStructure() error {
	for _, dir := range []string{
		".harness",
		".harness/artifacts",
		".harness/artifacts/design",
		".harness/artifacts/progress",
		".harness/skills",
		".opencode/skills",
		"skills",
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func assertStack(t *testing.T, stacks []StackSignal, want string) {
	t.Helper()
	for _, stack := range stacks {
		if stack.Name == want {
			return
		}
	}
	t.Fatalf("expected stack %s in %+v", want, stacks)
}
