package harness

import (
	"os"
	"testing"
)

func writeReviewArtifacts(t *testing.T, contract, qa, security, checklist string) {
	t.Helper()
	writeTestFile(t, "reports/contract-test-report.md", contract)
	writeTestFile(t, "reports/qa-report.md", qa)
	writeTestFile(t, "reports/security-review.md", security)
	writeTestFile(t, ReviewChecklistFile, checklist)
}

func validContractReport() string {
	return `# Contract Test Report

## Contract evidence:
Command output: shipwright contract check-mocks PASS; shipwright contract check-compliance PASS.

## Results
All contract checks pass.

## Issues found
- LOW: Contract tests are currently document-backed; track real API tests in hardening.
`
}

func validQAReport() string {
	return `# QA Report

## Test evidence:
Command output: go test ./... PASS.

## Test summary
Smoke checks passed.

## Issues found
- LOW: Manual exploratory coverage is limited.

## Recommendation
CONDITIONAL PASS
`
}

func validSecurityReview() string {
	return `# Security Review

## Security evidence:
Reviewed authentication, authorization, data exposure, and input validation notes.

## Findings
- MEDIUM: Rate limiting not implemented — Decision: accepted for MVP and tracked for hardening.

## Risk assessment
medium
`
}

func validChecklist() string {
	return `# Review Checklist

## Evidence gate

- [x] reports/contract-test-report.md includes Contract evidence
- [x] reports/qa-report.md includes Test evidence
- [x] reports/security-review.md includes Security evidence
- [x] No report is still placeholder-only

## Finding policy

- [x] CRITICAL findings: none open
- [x] MEDIUM findings: each has explicit decision
- [x] LOW findings: recorded and accepted for later tracking
`
}

func TestAssessReviewReportsBlocksPlaceholderContent(t *testing.T) {
	chdirTemp(t)
	writeReviewArtifacts(t,
		"# Contract\n\n## Contract evidence:\n(pending — command output)\n",
		validQAReport(),
		validSecurityReview(),
		validChecklist(),
	)

	assessment := AssessReviewReports()
	if !assessment.BlocksProgress() {
		t.Fatal("expected placeholder review content to block progress")
	}
}

func TestAssessReviewReportsBlocksCriticalFindings(t *testing.T) {
	chdirTemp(t)
	writeReviewArtifacts(t,
		validContractReport(),
		"# QA Report\n\n## Test evidence:\nCommand output: tests failed.\n\n## Issues found\n- CRITICAL: Checkout flow crashes.\n",
		validSecurityReview(),
		validChecklist(),
	)

	assessment := AssessReviewReports()
	if !assessment.BlocksProgress() {
		t.Fatal("expected critical finding to block progress")
	}
	if got, want := len(assessment.CriticalFindings), 1; got != want {
		t.Fatalf("critical findings = %d, want %d", got, want)
	}
}

func TestAssessReviewReportsBlocksUndecidedMediumFindings(t *testing.T) {
	chdirTemp(t)
	writeReviewArtifacts(t,
		validContractReport(),
		"# QA Report\n\n## Test evidence:\nCommand output: tests pass.\n\n## Issues found\n- MEDIUM: Missing edge-case test for invoice cancellation.\n",
		validSecurityReview(),
		validChecklist(),
	)

	assessment := AssessReviewReports()
	if !assessment.BlocksProgress() {
		t.Fatal("expected undecided medium finding to block progress")
	}
	if got, want := len(assessment.MediumPendingDecision), 1; got != want {
		t.Fatalf("medium pending = %d, want %d", got, want)
	}
}

func TestAssessReviewReportsAllowsLowAndDecidedMediumFindings(t *testing.T) {
	chdirTemp(t)
	writeReviewArtifacts(t, validContractReport(), validQAReport(), validSecurityReview(), validChecklist())

	assessment := AssessReviewReports()
	if assessment.BlocksProgress() {
		t.Fatalf("expected low + decided medium findings to pass, issues: %v", assessment.Issues)
	}
	if len(assessment.LowFindings) == 0 {
		t.Fatal("expected low findings to be recorded")
	}
	if len(assessment.MediumFindings) == 0 {
		t.Fatal("expected decided medium finding to be recorded")
	}
}

func TestAdvanceBlocksQASecurityReviewWithCriticalFindings(t *testing.T) {
	chdirTemp(t)
	writeReviewArtifacts(t,
		validContractReport(),
		validQAReport(),
		"# Security Review\n\n## Security evidence:\nReviewed auth.\n\n## Findings\n- HIGH: Admin endpoint allows unauthorized access.\n",
		validChecklist(),
	)

	state := NewState("Billing")
	state.CurrentPhase = StateQASecurityReview

	result := Advance(state)
	if result.Transitioned {
		t.Fatal("expected QA/security review to be blocked")
	}
	if result.BlockReason == "" {
		t.Fatal("expected block reason")
	}
}

func TestStartReviewArtifactsCreatesReviewFiles(t *testing.T) {
	chdirTemp(t)
	state := NewState("Billing")

	generated, _, errors := StartReviewArtifacts(state)
	if len(errors) > 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}
	if got, want := len(generated), 4; got != want {
		t.Fatalf("generated files = %d, want %d", got, want)
	}
	for _, path := range RequiredReviewArtifacts() {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}
}
