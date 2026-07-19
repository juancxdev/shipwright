package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCalibrateProjectDetectsNodeTypeScriptStack(t *testing.T) {
	withTempWorkingDir(t)
	writeProjectProfileTestFile(t, "package.json", `{
  "scripts": {
    "test": "vitest run",
    "build": "vite build",
    "lint": "eslint .",
    "dev": "vite --host 0.0.0.0"
  },
  "dependencies": {
    "react": "latest"
  },
  "devDependencies": {
    "typescript": "latest",
    "vite": "latest",
    "vitest": "latest"
  }
}`)
	writeProjectProfileTestFile(t, "tsconfig.json", `{}`)
	writeProjectProfileTestFile(t, "pnpm-lock.yaml", "lockfileVersion: '9.0'\n")
	if err := os.MkdirAll("src", 0755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}

	profile, err := CalibrateProject("webapp")
	if err != nil {
		t.Fatalf("CalibrateProject: %v", err)
	}

	assertStringSliceContains(t, profile.Languages, "TypeScript")
	assertStringSliceContains(t, profile.PackageManagers, "pnpm")
	assertCommand(t, profile.Commands.Test, "pnpm test")
	assertCommand(t, profile.Commands.Build, "pnpm run build")
	assertStack(t, profile.Stacks, "React")
	assertStack(t, profile.Stacks, "Vite")
	if !profile.TDD.Supported || profile.TDD.RecommendedMode != "strict" {
		t.Fatalf("tdd = %+v", profile.TDD)
	}
	if !profile.ExistingProject {
		t.Fatal("expected existing project")
	}
}

func TestCalibrateProjectDetectsGoProject(t *testing.T) {
	withTempWorkingDir(t)
	writeProjectProfileTestFile(t, "go.mod", "module example.com/app\n\ngo 1.24\n")
	if err := os.MkdirAll(filepath.Join("cmd", "api"), 0755); err != nil {
		t.Fatalf("mkdir cmd/api: %v", err)
	}
	if err := os.MkdirAll("internal", 0755); err != nil {
		t.Fatalf("mkdir internal: %v", err)
	}

	profile, err := CalibrateProject("api")
	if err != nil {
		t.Fatalf("CalibrateProject: %v", err)
	}

	assertStringSliceContains(t, profile.Languages, "Go")
	assertCommand(t, profile.Commands.Test, "go test ./...")
	assertCommand(t, profile.Commands.Build, "go build ./...")
	assertStack(t, profile.Stacks, "Go modules")
	assertStringSliceContains(t, profile.Conventions, "Go cmd/ entrypoint directory present")
	assertStringSliceContains(t, profile.Conventions, "Go internal/ package boundary present")
}

func TestSaveProjectProfileWritesJSONAndMarkdown(t *testing.T) {
	withTempWorkingDir(t)
	if err := CreateBaseStructure(); err != nil {
		t.Fatalf("CreateBaseStructure: %v", err)
	}
	writeProjectProfileTestFile(t, "go.mod", "module example.com/app\n\ngo 1.24\n")

	profile, err := CalibrateProject("api")
	if err != nil {
		t.Fatalf("CalibrateProject: %v", err)
	}
	if err := SaveProjectProfile(profile); err != nil {
		t.Fatalf("SaveProjectProfile: %v", err)
	}

	loaded, err := LoadProjectProfile()
	if err != nil {
		t.Fatalf("LoadProjectProfile: %v", err)
	}
	assertStringSliceContains(t, loaded.Languages, "Go")
	markdown, err := os.ReadFile(ProjectProfileMarkdown)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	if !strings.Contains(string(markdown), "Project Calibration Profile") || !strings.Contains(string(markdown), "go test ./...") {
		t.Fatalf("markdown missing calibration content:\n%s", string(markdown))
	}
}

func writeProjectProfileTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertStringSliceContains(t *testing.T, values []string, want string) {
	t.Helper()
	for _, value := range values {
		if value == want {
			return
		}
	}
	t.Fatalf("%q not found in %#v", want, values)
}

func assertCommand(t *testing.T, commands []DetectedCommand, want string) {
	t.Helper()
	for _, command := range commands {
		if command.Command == want {
			return
		}
	}
	t.Fatalf("command %q not found in %+v", want, commands)
}

func assertStack(t *testing.T, stacks []StackSignal, want string) {
	t.Helper()
	for _, stack := range stacks {
		if stack.Name == want {
			return
		}
	}
	t.Fatalf("stack %q not found in %+v", want, stacks)
}
