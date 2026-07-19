package cmd

import (
	"fmt"
	"strings"

	"shipwright/pkg/harness"
)

func Skills(args []string) {
	if len(args) == 0 {
		printSkillsUsage()
		return
	}
	switch args[0] {
	case "refresh":
		EnsureHarness()
		registry, err := harness.RefreshSkillRegistry()
		if err != nil {
			Fail(fmt.Sprintf("error refrescando skill registry: %s", err))
		}
		digests, err := harness.RefreshSkillDigestsFromRegistry(registry)
		if err != nil {
			Fail(fmt.Sprintf("error generando skill digests: %s", err))
		}
		PrintSuccess(fmt.Sprintf("Skill registry actualizado (%d skills)", len(registry.Skills)))
		PrintSuccess(fmt.Sprintf("Skill digests actualizados (%d agentes)", len(digests.Digests)))
		if len(registry.Warnings) > 0 || len(digests.Warnings) > 0 {
			fmt.Printf("Warnings: %d\n", len(registry.Warnings)+len(digests.Warnings))
		}
	case "status":
		EnsureHarness()
		registry, err := harness.LoadSkillRegistry()
		if err != nil {
			Fail("skill registry no encontrado. Ejecutá 'shipwright skills refresh'.")
		}
		printSkillRegistryStatus(registry)
	case "show":
		EnsureHarness()
		if len(args) < 2 {
			Fail("usage: shipwright skills show <name>")
		}
		registry, err := harness.LoadSkillRegistry()
		if err != nil {
			Fail("skill registry no encontrado. Ejecutá 'shipwright skills refresh'.")
		}
		skill := harness.FindSkill(registry, args[1])
		if skill == nil {
			Fail(fmt.Sprintf("skill %q no encontrada", args[1]))
		}
		printSkill(*skill)
	case "digest":
		EnsureHarness()
		digests, err := harness.LoadSkillDigests()
		if err != nil {
			Fail("skill digests no encontrados. Ejecutá 'shipwright skills refresh'.")
		}
		if len(args) >= 2 {
			digest := harness.FindSkillDigest(digests, args[1])
			if digest == nil {
				Fail(fmt.Sprintf("digest para agente %q no encontrado", args[1]))
			}
			printSkillDigest(*digest)
			return
		}
		printSkillDigestsStatus(digests)
	case "help", "-h", "--help":
		printSkillsUsage()
	default:
		Fail(fmt.Sprintf("unknown skills command: %s", args[0]))
	}
}

func printSkillsUsage() {
	fmt.Print(`Shipwright Skills

Usage:
  shipwright skills refresh        Scan skills and write registry + digests
  shipwright skills status         Show indexed skills and warnings
  shipwright skills show <name>    Show one indexed skill
  shipwright skills digest [agent] Show compact skill rules for all agents or one agent
`)
}

func printSkillRegistryStatus(registry *harness.SkillRegistry) {
	fmt.Println("Shipwright — Skill Registry")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println()
	fmt.Printf("Generated: %s\n", registry.GeneratedAt)
	fmt.Printf("Skills:    %d\n", len(registry.Skills))
	if len(registry.Sources) > 0 {
		fmt.Printf("Sources:   %s\n", strings.Join(registry.Sources, ", "))
	}
	fmt.Println()
	if len(registry.Skills) == 0 {
		fmt.Println("No skills indexed.")
	} else {
		for _, skill := range registry.Skills {
			desc := skill.Description
			if len(desc) > 90 {
				desc = desc[:87] + "..."
			}
			fmt.Printf("- %-24s %-16s %s\n", skill.Name, skill.Source, desc)
		}
	}
	if len(registry.Warnings) > 0 {
		fmt.Println()
		fmt.Println("Warnings:")
		for _, warning := range registry.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}
}

func printSkill(skill harness.SkillIndex) {
	fmt.Printf("Skill:      %s\n", skill.Name)
	fmt.Printf("Source:     %s\n", skill.Source)
	fmt.Printf("Path:       %s\n", skill.Path)
	if skill.Description != "" {
		fmt.Printf("Description: %s\n", skill.Description)
	}
	if len(skill.AppliesTo) > 0 {
		fmt.Printf("Applies to: %s\n", strings.Join(skill.AppliesTo, ", "))
	}
	if len(skill.Triggers) > 0 {
		fmt.Println("Triggers:")
		for _, trigger := range skill.Triggers {
			fmt.Printf("  - %s\n", trigger)
		}
	}
}

func printSkillDigestsStatus(digests *harness.SkillDigestSet) {
	fmt.Println("Shipwright — Skill Digests")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println()
	fmt.Printf("Generated: %s\n", digests.GeneratedAt)
	fmt.Printf("Agents:    %d\n", len(digests.Digests))
	fmt.Println()
	for _, digest := range digests.Digests {
		fmt.Printf("- %-24s skills=%d rules=%d\n", digest.Agent, len(digest.RelevantSkills), len(digest.CompactRules))
	}
	if len(digests.Warnings) > 0 {
		fmt.Println()
		fmt.Println("Warnings:")
		for _, warning := range digests.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}
}

func printSkillDigest(digest harness.AgentSkillDigest) {
	fmt.Printf("Agent: %s\n", digest.Agent)
	fmt.Println()
	if len(digest.RelevantSkills) > 0 {
		fmt.Println("Relevant skills:")
		for _, skill := range digest.RelevantSkills {
			fmt.Printf("  - %s — %s (%s)\n", skill.Name, skill.Reason, skill.Path)
		}
	} else {
		fmt.Println("Relevant skills: none")
	}
	if len(digest.CompactRules) > 0 {
		fmt.Println()
		fmt.Println("Compact rules:")
		for _, rule := range digest.CompactRules {
			fmt.Printf("  - %s\n", rule)
		}
	}
}
