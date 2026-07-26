package application

import (
	"strings"
	"testing"

	"shipwright/internal/design/domain"
)

func TestMissingBlockingChecksOnlyReturnsBlockingFailures(t *testing.T) {
	checks := []domain.GateCheck{
		{Gate: domain.GateBaselineCaptured, Pass: false, Blocking: true, Message: "missing screenshots"},
		{Gate: domain.GateProviderPublished, Pass: false, Blocking: false, Message: "optional warning"},
		{Gate: domain.GateTokenQuotaOK, Pass: true, Blocking: true, Message: "ok"},
	}
	missing := MissingBlockingChecks(checks)
	if len(missing) != 1 || !strings.Contains(missing[0], domain.GateBaselineCaptured) {
		t.Fatalf("missing = %+v", missing)
	}
}

func TestRenderGateSummaryMarksBlockedAndPass(t *testing.T) {
	summary := RenderGateSummary([]domain.GateCheck{
		{Gate: domain.GateBaselineCaptured, Pass: false, Blocking: true, Message: "missing"},
		{Gate: domain.GateTokenQuotaOK, Pass: true, Blocking: true, Message: "ok"},
	})
	if !strings.Contains(summary, "[BLOCKED]") || !strings.Contains(summary, "[PASS]") {
		t.Fatalf("summary missing marks:\n%s", summary)
	}
}
