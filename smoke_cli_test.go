package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"shipwright/cmd"
)

func TestCLISmokeHelpInitConfigValidateAndDoctor(t *testing.T) {
	tmp := t.TempDir()
	withWorkingDir(t, tmp)

	help := captureStdout(t, func() { cmd.PrintUsage() })
	for _, want := range []string{"shipwright config <subcommand>", "shipwright doctor [--json] [--fix]", "validate --json", "shipwright skills <subcommand>", "shipwright tdd <subcommand>"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help missing %q in:\n%s", want, help)
		}
	}

	initOutput := captureStdout(t, func() {
		cmd.Init([]string{
			"--executor", "opencode",
			"--reasoning-model", "openai/gpt-5.5",
			"--fast-model", "opencode-go/deepseek-v4-flash",
			"--agent-model", "technical-lead=openai/gpt-5.5",
		})
	})
	if !strings.Contains(initOutput, "Portable config creada") || !strings.Contains(initOutput, "Project calibration creada") || !strings.Contains(initOutput, "TDD policy creada") {
		t.Fatalf("init output missing portable config message:\n%s", initOutput)
	}
	for _, path := range []string{".harness/state.json", ".harness/config.json", ".harness/integrations.json", ".harness/project-profile.json", ".harness/project-profile.md", ".harness/tdd-policy.json", ".harness/tdd-policy.md", ".harness/skill-registry.json", ".harness/skill-registry.md", ".harness/skill-digests.json", ".harness/skill-digests.md", "AGENTS.md", ".harness/bin/shipwright", ".opencode/opencode.json", ".opencode/agents/product-owner.md"} {
		if _, err := os.Stat(filepath.Join(tmp, path)); err != nil {
			t.Fatalf("expected %s after init: %v", path, err)
		}
	}
	opencodeConfig, err := os.ReadFile(filepath.Join(tmp, ".opencode/opencode.json"))
	if err != nil {
		t.Fatalf("read opencode config: %v", err)
	}
	if !strings.Contains(string(opencodeConfig), "openai/gpt-5.5") || !strings.Contains(string(opencodeConfig), "opencode-go/deepseek-v4-flash") {
		t.Fatalf("opencode config missing custom models:\n%s", string(opencodeConfig))
	}

	validate := captureStdout(t, func() { cmd.Config([]string{"validate"}) })
	if !strings.Contains(validate, "Portable config is valid") {
		t.Fatalf("config validate output = %q", validate)
	}

	executorStatus := captureStdout(t, func() { cmd.Executor([]string{"status", "opencode"}) })
	if !strings.Contains(executorStatus, "Configured: yes") {
		t.Fatalf("executor status output = %q", executorStatus)
	}

	skillsStatus := captureStdout(t, func() { cmd.Skills([]string{"status"}) })
	if !strings.Contains(skillsStatus, "Skill Registry") || !strings.Contains(skillsStatus, "product-owner") {
		t.Fatalf("skills status output = %q", skillsStatus)
	}

	skillsDigest := captureStdout(t, func() { cmd.Skills([]string{"digest", "frontend-engineer"}) })
	if !strings.Contains(skillsDigest, "frontend-engineer") || !strings.Contains(skillsDigest, "Compact rules") {
		t.Fatalf("skills digest output = %q", skillsDigest)
	}

	tddStatus := captureStdout(t, func() { cmd.TDD([]string{"status"}) })
	if !strings.Contains(tddStatus, "Shipwright — TDD Status") || !strings.Contains(tddStatus, "Mode:") {
		t.Fatalf("tdd status output = %q", tddStatus)
	}

	doctor := captureStdout(t, func() { cmd.Doctor([]string{"--json"}) })
	if !strings.Contains(doctor, "\"summary\"") || !strings.Contains(doctor, "\"engram_health\"") {
		t.Fatalf("doctor json output = %q", doctor)
	}
}

func TestCLIInitDefaultsToOpenCodeExecutor(t *testing.T) {
	tmp := t.TempDir()
	withWorkingDir(t, tmp)

	initOutput := captureStdout(t, func() { cmd.Init(nil) })
	if !strings.Contains(initOutput, "Executor opencode generado") {
		t.Fatalf("init output should mention default opencode executor:\n%s", initOutput)
	}
	for _, path := range []string{"AGENTS.md", ".opencode/opencode.json", ".harness/bin/shipwright", ".harness/project-profile.md", ".harness/tdd-policy.md", ".harness/skill-registry.md", ".harness/skill-digests.md", ".harness/executor.json"} {
		if _, err := os.Stat(filepath.Join(tmp, path)); err != nil {
			t.Fatalf("expected %s after default init: %v", path, err)
		}
	}
}

func TestCLIInitAcceptsAIFlagAlias(t *testing.T) {
	tmp := t.TempDir()
	withWorkingDir(t, tmp)

	captureStdout(t, func() {
		cmd.Init([]string{
			"--ai", "opencode",
			"--reasoning-model", "opencode-go/deepseek-v4-flash",
			"--fast-model", "opencode-go/deepseek-v4-flash",
		})
	})
	opencodeConfig, err := os.ReadFile(filepath.Join(tmp, ".opencode/opencode.json"))
	if err != nil {
		t.Fatalf("read opencode config: %v", err)
	}
	if !strings.Contains(string(opencodeConfig), "opencode-go/deepseek-v4-flash") {
		t.Fatalf("--ai opencode init did not apply model flags:\n%s", string(opencodeConfig))
	}
}

func TestCLIVersionJSON(t *testing.T) {
	output := captureStdout(t, func() { cmd.VersionCommand([]string{"--json"}) })
	for _, want := range []string{`"version"`, `dev`, `"os"`, `"arch"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("version json missing %q in:\n%s", want, output)
		}
	}
}

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	_ = w.Close()
	os.Stdout = old
	return <-done
}
