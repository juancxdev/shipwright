package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"shipwright/pkg/harness"
)

func Agents(args []string) {
	EnsureHarness()

	if len(args) == 0 {
		PrintUsage()
		Fail("usage: shipwright agents <list|show|active|run>")
	}

	subcommand := args[0]
	rest := args[1:]

	switch subcommand {
	case "list":
		agentsList(rest)
	case "show":
		agentsShow(rest)
	case "active":
		agentsActive(rest)
	case "run":
		agentsRun(rest)
	default:
		Fail(fmt.Sprintf("unknown agents subcommand: %s\n\nValid: list | show | active | run", subcommand))
	}
}

func agentsList(args []string) {
	fmt.Println("Shipwright — Agents")
	fmt.Println("======================")
	fmt.Println()

	for _, skill := range harness.AllAgentSkills() {
		lines := strings.Split(skill.Content, "\n")
		desc := ""
		for _, line := range lines {
			if strings.HasPrefix(line, "description:") {
				desc = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				desc = strings.Trim(desc, "\"")
				break
			}
		}
		agent := harness.GetAgent(skill.Name)
		phases := ""
		if agent != nil {
			phases = agentPhases(skill.Name)
		}
		fmt.Printf("  %-22s  %-20s  %s\n", skill.Name, phases, truncate(desc, 50))
	}

	fmt.Println()

	state, _ := harness.LoadState()
	if state != nil {
		active := harness.ActiveAgentForPhase(state.CurrentPhase)
		if active != nil {
			fmt.Printf("Active agent: %s (phase: %s)\n", active.Name, state.CurrentPhase)
		} else {
			fmt.Printf("Active agent: none (phase: %s)\n", state.CurrentPhase)
		}
	}

	fmt.Println()
	fmt.Println("Usage: shipwright agents show <name>   — full SKILL.md")
	fmt.Println("       shipwright agents active        — active agent for current phase")
	fmt.Println("       shipwright agents run <name>    — output SKILL.md for AI execution")
}

func agentsShow(args []string) {
	if len(args) == 0 {
		Fail("usage: shipwright agents show <name>\n\nAvailable: product-owner | project-manager | technical-lead | ui-ux-designer | frontend-engineer | backend-engineer | qa-security-reviewer")
	}

	name := args[0]
	skill := harness.GetAgentSkill(name)
	if skill == nil {
		Fail(fmt.Sprintf("unknown agent: %s\n\nAvailable: product-owner | project-manager | technical-lead | ui-ux-designer | frontend-engineer | backend-engineer | qa-security-reviewer", name))
	}

	path := filepath.Join(".harness/agents", name+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Print(skill.Content)
	} else {
		fmt.Print(string(data))
	}
}

func agentsActive(args []string) {
	state, err := harness.LoadState()
	if err != nil {
		Fail(err.Error())
	}

	active := harness.ActiveAgentForPhase(state.CurrentPhase)
	secondary := harness.SecondaryAgentForPhase(state.CurrentPhase)

	fmt.Println("Shipwright — Active Agent")
	fmt.Println("============================")
	fmt.Println()

	if active == nil {
		fmt.Printf("Phase: %s\n", state.CurrentPhase)
		fmt.Println("Active agent: none")
		return
	}

	fmt.Printf("Phase:           %s\n", state.CurrentPhase)
	fmt.Printf("Primary agent:   %s\n", active.Name)
	if secondary != nil {
		fmt.Printf("Secondary agent: %s\n", secondary.Name)
	}
	fmt.Println()

	fmt.Println("Purpose:")
	fmt.Printf("  %s\n\n", active.Purpose)

	fmt.Println("Can modify:")
	for _, m := range active.CanModify {
		exists := harness.ArtifactExists(m)
		mark := " "
		if exists {
			mark = "✓"
		}
		fmt.Printf("  [%s] %s\n", mark, m)
	}
	fmt.Println()

	fmt.Println("Can read:")
	for _, r := range active.CanRead {
		exists := harness.ArtifactExists(r)
		mark := " "
		if exists {
			mark = "✓"
		}
		fmt.Printf("  [%s] %s\n", mark, r)
	}
	fmt.Println()

	fmt.Println("What to Do:")
	for i, step := range active.Steps {
		fmt.Printf("  %d. %s\n", i+1, step.Title)
	}
	fmt.Println()

	fmt.Println("Done criteria:")
	for i, c := range active.DoneCriteria {
		fmt.Printf("  %d. %s\n", i+1, c)
	}
	fmt.Println()

	fmt.Println("Never:")
	for _, n := range active.Never {
		fmt.Printf("  ✗ %s\n", n)
	}

	fmt.Println()
	handoffCount := harness.CountHandoffs()
	fmt.Printf("Handoff log: %s (%d entries)\n", harness.HandoffLogFile, handoffCount)

	fmt.Println()
	fmt.Printf("To get full SKILL.md for AI execution:\n  shipwright agents run %s\n", active.Name)
}

func agentsRun(args []string) {
	if len(args) == 0 {
		Fail("usage: shipwright agents run <name>\n\nAvailable: product-owner | project-manager | technical-lead | ui-ux-designer | frontend-engineer | backend-engineer | qa-security-reviewer")
	}

	name := args[0]
	skill := harness.GetAgentSkill(name)
	if skill == nil {
		Fail(fmt.Sprintf("unknown agent: %s\n\nAvailable: product-owner | project-manager | technical-lead | ui-ux-designer | frontend-engineer | backend-engineer | qa-security-reviewer", name))
	}

	path := filepath.Join(".harness/agents", name+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Print(skill.Content)
	} else {
		fmt.Print(string(data))
	}

	fmt.Println()
	fmt.Println("---")
	fmt.Println("To execute this agent:")
	fmt.Println("1. Read _shared/agent-common.md for the shared protocol")
	fmt.Println("2. Follow ALL steps in this SKILL.md in order")
	fmt.Println("3. Respect ALL Hard Rules and NEVER rules")
	fmt.Println("4. Write artifacts to the paths listed in 'Can Modify'")
	fmt.Println("5. Use the exact output templates provided in each step")
	fmt.Println("6. Return the structured envelope to the orchestrator")
	fmt.Println("7. Run: shipwright next  (to advance to the next phase)")
}

func agentPhases(name string) string {
	phases := []string{}
	for _, state := range harness.AllStates {
		active := harness.ActiveAgentForPhase(state)
		if active != nil && active.Name == name {
			phases = append(phases, state)
		}
	}
	if len(phases) <= 2 {
		return strings.Join(phases, ", ")
	}
	return fmt.Sprintf("%s...%s", phases[0], phases[len(phases)-1])
}
