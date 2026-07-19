package cmd

import (
	"encoding/json"
	"fmt"

	"shipwright/pkg/harness"
)

func Doctor(args []string) {
	EnsureHarness()

	jsonOutput := false
	fix := false
	for _, arg := range args {
		switch arg {
		case "--json":
			jsonOutput = true
		case "--fix":
			fix = true
		case "-h", "--help":
			fmt.Println("usage: shipwright doctor [--json] [--fix]")
			return
		default:
			Fail(fmt.Sprintf("unknown doctor option: %s\n\nValid: --json | --fix", arg))
		}
	}

	var fixResult *harness.DoctorFixResult
	if fix {
		var err error
		fixResult, err = harness.ApplyDoctorFixes(harness.RealSystemProbe{})
		if err != nil {
			Fail(fmt.Sprintf("doctor fix failed: %s", err))
		}
	}

	report, err := harness.RunDoctor(harness.RealSystemProbe{})
	if err != nil {
		Fail(fmt.Sprintf("doctor failed: %s", err))
	}

	if jsonOutput {
		payload := any(report)
		if fixResult != nil {
			payload = struct {
				Fix    *harness.DoctorFixResult `json:"fix"`
				Report *harness.DoctorReport    `json:"report"`
			}{Fix: fixResult, Report: report}
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			Fail(fmt.Sprintf("cannot render doctor report: %s", err))
		}
		fmt.Println(string(data))
		if report.Summary.Errors > 0 {
			Exit(2)
		}
		return
	}

	if fixResult != nil {
		printDoctorFixResult(fixResult)
	}
	printDoctorReport(report)
	if report.Summary.Errors > 0 {
		Exit(2)
	}
}

func printDoctorFixResult(result *harness.DoctorFixResult) {
	fmt.Println("Shipwright — Doctor Fixes")
	fmt.Println("===========================")
	if result == nil || !result.Applied {
		fmt.Println("No fixes applied.")
		fmt.Println()
		return
	}
	for _, action := range result.Actions {
		fmt.Printf("✓ %s\n", action)
	}
	if result.Backup != "" {
		fmt.Printf("Backup: %s\n", result.Backup)
	}
	fmt.Println()
}

func printDoctorReport(report *harness.DoctorReport) {
	fmt.Println("Shipwright — Doctor")
	fmt.Println("======================")
	fmt.Println()
	fmt.Println("Platform:")
	fmt.Printf("  os:   %s\n", report.Platform.OS)
	fmt.Printf("  arch: %s\n", report.Platform.Arch)
	fmt.Printf("  ci:   %s\n", boolYesNo(report.Platform.IsCI))
	fmt.Println()

	fmt.Println("Config:")
	fmt.Printf("  file:   %s\n", report.ConfigFile)
	fmt.Printf("  exists: %s\n", boolYesNo(report.ConfigExists))
	fmt.Printf("  loaded: %s\n", boolYesNo(report.ConfigLoaded))
	fmt.Println()

	fmt.Println("Checks:")
	for _, check := range report.Checks {
		fmt.Printf("  [%s] %s — %s\n", check.Severity, check.Title, check.Status)
		if check.Detail != "" {
			fmt.Printf("       detail: %s\n", check.Detail)
		}
		if check.Action != "" {
			fmt.Printf("       action: %s\n", check.Action)
		}
	}
	fmt.Println()

	fmt.Println("Integrations:")
	printDoctorDetection("Engram", report.Engram)
	printDoctorHealth("Engram health", report.EngramHealth)
	printDoctorDetection("OpenPencil", report.OpenPencil)
	printDoctorHealth("OpenPencil health", report.OpenPencilHealth)
	fmt.Println()

	if len(report.Actions) > 0 {
		fmt.Println("Recommended actions:")
		for i, action := range report.Actions {
			fmt.Printf("  %d. %s\n", i+1, action)
		}
		fmt.Println()
	}

	fmt.Println("Summary:")
	fmt.Printf("  ok:       %d\n", report.Summary.OK)
	fmt.Printf("  info:     %d\n", report.Summary.Info)
	fmt.Printf("  warnings: %d\n", report.Summary.Warnings)
	fmt.Printf("  errors:   %d\n", report.Summary.Errors)
	if report.Summary.Healthy {
		PrintSuccess("Doctor finished without blocking errors.")
	} else {
		PrintError("Doctor found blocking errors.")
	}
}

func printDoctorHealth(label string, result harness.HealthResult) {
	fmt.Printf("    %s:\n", label)
	fmt.Printf("      status:  %s\n", result.Status)
	fmt.Printf("      checked: %s\n", boolYesNo(result.Checked))
	fmt.Printf("      healthy: %s\n", boolYesNo(result.Healthy))
	if result.Endpoint != "" {
		fmt.Printf("      target:  %s\n", result.Endpoint)
	}
	if result.LatencyMS > 0 {
		fmt.Printf("      latency: %dms\n", result.LatencyMS)
	}
	if result.Detail != "" {
		fmt.Printf("      detail:  %s\n", result.Detail)
	}
}

func printDoctorDetection(label string, result harness.DetectionResult) {
	fmt.Printf("  %s:\n", label)
	fmt.Printf("    status:    %s\n", result.Status)
	fmt.Printf("    installed: %s\n", boolYesNo(result.Installed))
	fmt.Printf("    available: %s\n", boolYesNo(result.Available))
	if result.Path != "" {
		if result.PathKind != "" {
			fmt.Printf("    path:      %s: %s\n", result.PathKind, result.Path)
		} else {
			fmt.Printf("    path:      %s\n", result.Path)
		}
	}
	if result.Reason != "" {
		fmt.Printf("    reason:    %s\n", result.Reason)
	}
	if result.Fallback != "" {
		fmt.Printf("    fallback:  %s\n", result.Fallback)
	}
}
