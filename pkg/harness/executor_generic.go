package harness

import "strings"

type GenericExecutorAdapter struct{}

func (GenericExecutorAdapter) Name() string { return ExecutorGeneric }

func (GenericExecutorAdapter) Description() string {
	return "Generic AI-agent bootstrap: AGENTS.md with Shipwright lifecycle rules."
}

func (GenericExecutorAdapter) Generate() (*ExecutorGenerateResult, error) {
	result := &ExecutorGenerateResult{Name: ExecutorGeneric}
	if err := writeTrackedFile("AGENTS.md", genericAgentsMD(), result); err != nil {
		return nil, err
	}
	result.Message = "Generic executor instructions generated in AGENTS.md."
	return result, nil
}

func (GenericExecutorAdapter) Status() (*ExecutorStatus, error) {
	return requiredStatus(ExecutorGeneric, []string{"AGENTS.md"}), nil
}

func genericAgentsMD() string {
	var sb strings.Builder
	sb.WriteString("# Shipwright Project Instructions\n\n")
	sb.WriteString("This project is managed by Shipwright. Shipwright is the source of truth for lifecycle, gates, roles, contracts, reviews, and evidence.\n\n")
	sb.WriteString("## Mandatory workflow\n\n")
	sb.WriteString("1. Read `.harness/project-profile.md`, `.harness/tdd-policy.md`, `.harness/skill-registry.md`, and `.harness/skill-digests.md` before making technical assumptions.\n")
	sb.WriteString("2. Run `shipwright status` before making changes.\n")
	sb.WriteString("3. Run `shipwright agents active` to identify the active role.\n")
	sb.WriteString("4. Run `shipwright agents run <agent-name>` and follow that role's instructions.\n")
	sb.WriteString("5. Do not advance phases manually; use `shipwright next`.\n")
	sb.WriteString("6. Do not approve gates yourself; user approvals must use `shipwright approve <gate>`.\n")
	sb.WriteString("7. Do not mark work finished without evidence in `reports/` when the phase requires it.\n")
	sb.WriteString("8. If `.harness/tdd-policy.md` says `strict`, record executed test evidence before integration.\n")
	sb.WriteString("9. Run `shipwright doctor` if environment or integration behavior is unclear.\n\n")
	sb.WriteString("## Contract-first rules\n\n")
	sb.WriteString("- Backend must implement `contracts/openapi.yaml`.\n")
	sb.WriteString("- Frontend must preserve mock mode and real API mode.\n")
	sb.WriteString("- Contract changes require explicit technical approval/change request.\n\n")
	sb.WriteString("## Safety\n\n")
	sb.WriteString("- Shipwright orchestrates. The AI coding agent executes.\n")
	sb.WriteString("- If scope, target, or phase is ambiguous, stop and ask. Do not guess.\n")
	return sb.String()
}
