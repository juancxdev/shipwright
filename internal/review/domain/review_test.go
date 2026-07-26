package domain

import "testing"

func TestAssessmentGateRulesBlockCriticalFindings(t *testing.T) {
	assessment := &Assessment{
		HasRealEvidence:  true,
		CriticalFindings: []Finding{{Severity: SeverityCritical}},
	}

	assessment.ApplyGateRules()

	if !assessment.BlocksProgress() {
		t.Fatalf("expected critical findings to block progress")
	}
}

func TestAssessmentGateRulesAllowLowFindingsWithWarning(t *testing.T) {
	assessment := &Assessment{
		HasRealEvidence: true,
		LowFindings:     []Finding{{Severity: SeverityLow}},
	}

	assessment.ApplyGateRules()

	if assessment.BlocksProgress() {
		t.Fatalf("expected low findings not to block progress: %v", assessment.Issues)
	}
	if len(assessment.Warnings) != 1 {
		t.Fatalf("expected one warning, got %d", len(assessment.Warnings))
	}
}
