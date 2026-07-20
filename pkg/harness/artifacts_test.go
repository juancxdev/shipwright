package harness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateBaseStructureKeepsLifecycleArtifactsUnderHarness(t *testing.T) {
	chdirTemp(t)

	if err := CreateBaseStructure(); err != nil {
		t.Fatalf("CreateBaseStructure: %v", err)
	}

	for _, dir := range artifactDirs {
		path := filepath.Join(ArtifactRoot, filepath.FromSlash(dir))
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			t.Fatalf("expected artifact dir %s: info=%v err=%v", path, info, err)
		}
	}

	for _, legacy := range []string{"product", "project", "design", "architecture", "contracts", "backlog", "sdd", "knowledge", "progress", "reports"} {
		if _, err := os.Stat(legacy); err == nil {
			t.Fatalf("legacy root artifact dir should not be created: %s", legacy)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat legacy dir %s: %v", legacy, err)
		}
	}
}

func TestArtifactPathMapsLegacyLifecyclePaths(t *testing.T) {
	got := ArtifactPath("product/context.md")
	want := filepath.Join(ArtifactRoot, "product", "context.md")
	if got != want {
		t.Fatalf("ArtifactPath = %q, want %q", got, want)
	}

	already := filepath.Join(ArtifactRoot, "product", "context.md")
	if got := ArtifactPath(already); got != already {
		t.Fatalf("ArtifactPath should keep migrated path, got %q", got)
	}
}
