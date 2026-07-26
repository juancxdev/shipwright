package domain

import "testing"

func TestResolvePathMapsLifecycleArtifactDirs(t *testing.T) {
	got := ResolvePath(DefaultLayout(), "design/prototype.md")
	want := ".harness/artifacts/design/prototype.md"
	if got != want {
		t.Fatalf("path = %s, want %s", got, want)
	}
}

func TestResolvePathLeavesHarnessPathsUntouched(t *testing.T) {
	path := ".harness/config.json"
	if got := ResolvePath(DefaultLayout(), path); got != path {
		t.Fatalf("path = %s, want %s", got, path)
	}
}
