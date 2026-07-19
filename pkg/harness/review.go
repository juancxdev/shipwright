package harness

import (
	"fmt"
	"os"
	"strings"
)

const ReviewChecklistFile = "reports/review-checklist.md"

type FindingSeverity string

const (
	SeverityCritical FindingSeverity = "critical"
	SeverityMedium   FindingSeverity = "medium"
	SeverityLow      FindingSeverity = "low"
)

type ReviewFinding struct {
	Source   string
	Severity FindingSeverity
	Line     string
	Decided  bool
}

type ReviewAssessment struct {
	ReportsPresent        bool
	HasRealEvidence       bool
	CriticalFindings      []ReviewFinding
	MediumFindings        []ReviewFinding
	MediumPendingDecision []ReviewFinding
	LowFindings           []ReviewFinding
	Issues                []string
	Warnings              []string
}

func RequiredContractReviewArtifacts() []string {
	return []string{
		"reports/contract-test-report.md",
		ReviewChecklistFile,
	}
}

func RequiredReviewArtifacts() []string {
	return []string{
		"reports/contract-test-report.md",
		"reports/qa-report.md",
		"reports/security-review.md",
		ReviewChecklistFile,
	}
}

func StartReviewArtifacts(state *State) ([]string, []string, []string) {
	data := TemplateDataFromState(state)
	paths := RequiredReviewArtifacts()

	var generated []string
	var skipped []string
	var errors []string

	for _, path := range paths {
		if ArtifactExists(path) {
			skipped = append(skipped, path)
			continue
		}

		content, err := GenerateArtifact(path, data)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", path, err))
			continue
		}
		if err := WriteFile(path, content); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", path, err))
			continue
		}
		generated = append(generated, path)
	}

	return generated, skipped, errors
}

func AssessContractReviewReports() *ReviewAssessment {
	return assessReviewPaths(RequiredContractReviewArtifacts())
}

func AssessReviewReports() *ReviewAssessment {
	return assessReviewPaths(RequiredReviewArtifacts())
}

func assessReviewPaths(paths []string) *ReviewAssessment {
	assessment := &ReviewAssessment{}

	missing := CheckArtifacts(paths)
	if len(missing) > 0 {
		assessment.Issues = append(assessment.Issues, "missing review artifacts: "+strings.Join(missing, ", "))
		return assessment
	}
	assessment.ReportsPresent = true

	for _, path := range paths {
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			assessment.Issues = append(assessment.Issues, fmt.Sprintf("cannot read %s: %s", path, err))
			continue
		}

		content := string(contentBytes)
		if path != ReviewChecklistFile && isPlaceholderReviewContent(content) {
			assessment.Issues = append(assessment.Issues, fmt.Sprintf("%s still contains placeholder/no-evidence content", path))
		}
		if hasEvidenceMarker(content) {
			assessment.HasRealEvidence = true
		}
		if path == ReviewChecklistFile {
			continue
		}

		for _, finding := range extractFindings(path, content) {
			switch finding.Severity {
			case SeverityCritical:
				assessment.CriticalFindings = append(assessment.CriticalFindings, finding)
			case SeverityMedium:
				assessment.MediumFindings = append(assessment.MediumFindings, finding)
				if !finding.Decided {
					assessment.MediumPendingDecision = append(assessment.MediumPendingDecision, finding)
				}
			case SeverityLow:
				assessment.LowFindings = append(assessment.LowFindings, finding)
			}
		}
	}

	if !assessment.HasRealEvidence {
		assessment.Issues = append(assessment.Issues, "review reports do not include evidence markers (Evidence:, Test evidence:, Security evidence:, or Contract evidence:)")
	}
	if len(assessment.CriticalFindings) > 0 {
		assessment.Issues = append(assessment.Issues, fmt.Sprintf("%d critical/high finding(s) block user acceptance", len(assessment.CriticalFindings)))
	}
	if len(assessment.MediumPendingDecision) > 0 {
		assessment.Issues = append(assessment.Issues, fmt.Sprintf("%d medium finding(s) require explicit decision", len(assessment.MediumPendingDecision)))
	}
	if len(assessment.LowFindings) > 0 && len(assessment.Issues) == 0 {
		assessment.Warnings = append(assessment.Warnings, fmt.Sprintf("%d low finding(s) recorded; they do not block", len(assessment.LowFindings)))
	}

	return assessment
}

func (a *ReviewAssessment) BlocksProgress() bool {
	return len(a.Issues) > 0
}

func FormatReviewAssessment(a *ReviewAssessment) string {
	var sb strings.Builder

	sb.WriteString("Shipwright — Review Status\n")
	sb.WriteString("============================\n\n")
	sb.WriteString(fmt.Sprintf("Reports present:       %s\n", formatYesNo(a.ReportsPresent)))
	sb.WriteString(fmt.Sprintf("Evidence present:      %s\n", formatYesNo(a.HasRealEvidence)))
	sb.WriteString(fmt.Sprintf("Critical findings:     %d\n", len(a.CriticalFindings)))
	sb.WriteString(fmt.Sprintf("Medium findings:       %d\n", len(a.MediumFindings)))
	sb.WriteString(fmt.Sprintf("Medium pending:        %d\n", len(a.MediumPendingDecision)))
	sb.WriteString(fmt.Sprintf("Low findings:          %d\n", len(a.LowFindings)))

	if len(a.Issues) > 0 {
		sb.WriteString("\nBlocking issues:\n")
		for _, issue := range a.Issues {
			sb.WriteString(fmt.Sprintf("  ✗ %s\n", issue))
		}
	}

	if len(a.Warnings) > 0 {
		sb.WriteString("\nWarnings:\n")
		for _, warning := range a.Warnings {
			sb.WriteString(fmt.Sprintf("  ⚠ %s\n", warning))
		}
	}

	writeFindings := func(title string, findings []ReviewFinding) {
		if len(findings) == 0 {
			return
		}
		sb.WriteString("\n" + title + ":\n")
		for _, finding := range findings {
			decision := ""
			if finding.Severity == SeverityMedium {
				decision = fmt.Sprintf(" decided=%s", formatYesNo(finding.Decided))
			}
			sb.WriteString(fmt.Sprintf("  - [%s] %s%s — %s\n", finding.Severity, finding.Source, decision, finding.Line))
		}
	}

	writeFindings("Critical/high findings", a.CriticalFindings)
	writeFindings("Medium findings", a.MediumFindings)
	writeFindings("Low findings", a.LowFindings)

	if a.BlocksProgress() {
		sb.WriteString("\n✗ Review gate: BLOCKED\n")
	} else {
		sb.WriteString("\n✓ Review gate: PASS\n")
	}

	return sb.String()
}

func ContractReviewBlockReason() string {
	assessment := AssessContractReviewReports()
	if !assessment.BlocksProgress() {
		return ""
	}
	return FormatReviewAssessment(assessment)
}

func ReviewBlockReason() string {
	assessment := AssessReviewReports()
	if !assessment.BlocksProgress() {
		return ""
	}
	return FormatReviewAssessment(assessment)
}

func isPlaceholderReviewContent(content string) bool {
	lower := strings.ToLower(content)
	return strings.Contains(lower, "placeholder") ||
		strings.Contains(lower, "no real") ||
		strings.Contains(lower, "(pending") ||
		strings.Contains(lower, "(none")
}

func hasEvidenceMarker(content string) bool {
	lower := strings.ToLower(content)
	markers := []string{
		"evidence:",
		"test evidence:",
		"security evidence:",
		"contract evidence:",
		"command output:",
		"verification:",
	}
	for _, marker := range markers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func extractFindings(source, content string) []ReviewFinding {
	var findings []ReviewFinding

	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || !looksLikeFindingLine(line) {
			continue
		}
		lower := strings.ToLower(line)

		severity, ok := findingSeverity(lower)
		if !ok {
			continue
		}

		findings = append(findings, ReviewFinding{
			Source:   source,
			Severity: severity,
			Line:     line,
			Decided:  mediumDecisionPresent(lower),
		})
	}

	return findings
}

func findingSeverity(lowerLine string) (FindingSeverity, bool) {
	if strings.Contains(lowerLine, "critical") || strings.Contains(lowerLine, "high") || strings.Contains(lowerLine, "bloqueante") {
		return SeverityCritical, true
	}
	if strings.Contains(lowerLine, "medium") || strings.Contains(lowerLine, "medio") || strings.Contains(lowerLine, "major") {
		return SeverityMedium, true
	}
	if strings.Contains(lowerLine, "low") || strings.Contains(lowerLine, "bajo") || strings.Contains(lowerLine, "minor") {
		return SeverityLow, true
	}
	return "", false
}

func mediumDecisionPresent(lowerLine string) bool {
	decisionMarkers := []string{
		"decision:",
		"decisión:",
		"accepted",
		"aceptado",
		"approved",
		"aprobado",
		"waived",
		"asumido",
		"deferred",
		"postergado",
		"mitigated",
		"mitigado",
		"[x]",
	}
	for _, marker := range decisionMarkers {
		if strings.Contains(lowerLine, marker) {
			return true
		}
	}
	return false
}

func looksLikeFindingLine(line string) bool {
	return strings.HasPrefix(line, "-") ||
		strings.HasPrefix(line, "*") ||
		strings.HasPrefix(line, "[")
}
