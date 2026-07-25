package harness

import (
	"os"
	"strings"
	"testing"
)

func TestSaveLocalSecretWritesIgnoredMachineLocalFile(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	if err := SaveLocalSecret("STITCH_API_KEY", "abc123"); err != nil {
		t.Fatalf("SaveLocalSecret error = %v", err)
	}

	secrets := LoadLocalSecrets()
	if secrets["STITCH_API_KEY"] != "abc123" {
		t.Fatalf("secret = %q", secrets["STITCH_API_KEY"])
	}

	ignoreBytes, err := os.ReadFile(HarnessGitignore)
	if err != nil {
		t.Fatalf("reading gitignore: %v", err)
	}
	if !strings.Contains(string(ignoreBytes), "secrets.local.env") {
		t.Fatalf("expected .harness/.gitignore to ignore secrets.local.env, got %q", string(ignoreBytes))
	}

	info, err := os.Stat(LocalSecretsFile)
	if err != nil {
		t.Fatalf("stat secrets file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("secret permissions = %v, want 0600", info.Mode().Perm())
	}
}

func TestDetectStitchReadsLocalSecret(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	if err := SaveLocalSecret("STITCH_API_KEY", "abc123"); err != nil {
		t.Fatalf("SaveLocalSecret error = %v", err)
	}

	result := DetectStitch(fakeProbe{goos: "linux"})
	if !result.Available || result.Status != DetectionAvailable {
		t.Fatalf("DetectStitch = %+v", result)
	}
}
