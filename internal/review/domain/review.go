package domain

import "fmt"

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

type Finding struct {
	Source   string
	Severity Severity
	Line     string
	Decided  bool
}

type Assessment struct {
	ReportsPresent        bool
	HasRealEvidence       bool
	CriticalFindings      []Finding
	MediumFindings        []Finding
	MediumPendingDecision []Finding
	LowFindings           []Finding
	Issues                []string
	Warnings              []string
}

func (a *Assessment) BlocksProgress() bool {
	return a != nil && len(a.Issues) > 0
}

func (a *Assessment) ApplyGateRules() {
	if a == nil {
		return
	}
	if !a.HasRealEvidence {
		a.Issues = append(a.Issues, "review reports do not include evidence markers (Evidence:, Test evidence:, Security evidence:, or Contract evidence:)")
	}
	if len(a.CriticalFindings) > 0 {
		a.Issues = append(a.Issues, fmt.Sprintf("%d critical/high finding(s) block user acceptance", len(a.CriticalFindings)))
	}
	if len(a.MediumPendingDecision) > 0 {
		a.Issues = append(a.Issues, fmt.Sprintf("%d medium finding(s) require explicit decision", len(a.MediumPendingDecision)))
	}
	if len(a.LowFindings) > 0 && len(a.Issues) == 0 {
		a.Warnings = append(a.Warnings, fmt.Sprintf("%d low finding(s) recorded; they do not block", len(a.LowFindings)))
	}
}
