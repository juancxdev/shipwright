package application

import (
	"fmt"
	"strings"

	designdomain "shipwright/internal/design/domain"
)

func RenderGateSummary(checks []designdomain.GateCheck) string {
	var sb strings.Builder
	for _, check := range checks {
		mark := "PASS"
		if !check.Pass {
			mark = "BLOCKED"
			if !check.Blocking {
				mark = "WARN"
			}
		}
		sb.WriteString(fmt.Sprintf("- [%s] %s — %s\n", mark, check.Gate, check.Message))
	}
	return sb.String()
}

func MissingBlockingChecks(checks []designdomain.GateCheck) []string {
	var missing []string
	for _, check := range checks {
		if check.Blocking && !check.Pass {
			missing = append(missing, fmt.Sprintf("%s — %s", check.Gate, check.Message))
		}
	}
	return missing
}
