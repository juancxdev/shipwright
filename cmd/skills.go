package cmd

import (
	"encoding/json"
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
		assignments, err := harness.RefreshSkillAssignmentsFromRegistry(registry)
		if err != nil {
			Fail(fmt.Sprintf("error generando skill assignments: %s", err))
		}
		digests, err := harness.RefreshSkillDigestsFromRegistry(registry)
		if err != nil {
			Fail(fmt.Sprintf("error generando skill digests: %s", err))
		}
		PrintSuccess(fmt.Sprintf("Skill registry actualizado (%d skills)", len(registry.Skills)))
		PrintSuccess(fmt.Sprintf("Skill assignments actualizados (%d tecnologías, %d recomendaciones)", len(assignments.Technologies), len(assignments.Skills)))
		PrintSuccess(fmt.Sprintf("Skill digests actualizados (%d agentes)", len(digests.Digests)))
		if len(registry.Warnings) > 0 || len(assignments.Warnings) > 0 || len(digests.Warnings) > 0 {
			fmt.Printf("Warnings: %d\n", len(registry.Warnings)+len(assignments.Warnings)+len(digests.Warnings))
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
	case "assign":
		EnsureHarness()
		assignments, err := harness.RefreshSkillAssignments()
		if err != nil {
			Fail(fmt.Sprintf("error generando skill assignments: %s", err))
		}
		if len(args) >= 2 && args[1] == "--json" {
			printSkillAssignmentsJSON(assignments)
			return
		}
		printSkillAssignmentsStatus(assignments)
	case "import":
		EnsureHarness()
		if len(args) < 2 {
			Fail("usage: shipwright skills import autoskills [--json]")
		}
		switch args[1] {
		case "autoskills":
			result, err := harness.ImportAutoSkillsToOpenCode()
			if err != nil {
				Fail(fmt.Sprintf("error importando autoskills: %s", err))
			}
			registry, err := harness.RefreshSkillRegistry()
			if err != nil {
				Fail(fmt.Sprintf("error refrescando skill registry: %s", err))
			}
			assignments, err := harness.RefreshSkillAssignmentsFromRegistry(registry)
			if err != nil {
				Fail(fmt.Sprintf("error generando skill assignments: %s", err))
			}
			digests, err := harness.RefreshSkillDigestsFromRegistry(registry)
			if err != nil {
				Fail(fmt.Sprintf("error generando skill digests: %s", err))
			}
			if len(args) >= 3 && args[2] == "--json" {
				printAutoSkillsImportJSON(result)
				return
			}
			PrintSuccess(fmt.Sprintf("Autoskills importadas a .opencode/skills (%d)", len(result.Imported)))
			if len(result.Imported) > 0 {
				fmt.Printf("Skills: %s\n", strings.Join(result.Imported, ", "))
			}
			if len(result.Skipped) > 0 {
				fmt.Printf("Skipped: %s\n", strings.Join(result.Skipped, ", "))
			}
			PrintSuccess(fmt.Sprintf("Skill registry actualizado (%d skills)", len(registry.Skills)))
			PrintSuccess(fmt.Sprintf("Skill assignments actualizados (%d tecnologías, %d recomendaciones)", len(assignments.Technologies), len(assignments.Skills)))
			PrintSuccess(fmt.Sprintf("Skill digests actualizados (%d agentes)", len(digests.Digests)))
		default:
			Fail(fmt.Sprintf("unknown skills provider: %s", args[1]))
		}
	case "providers":
		EnsureHarness()
		printSkillProviders()
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
  shipwright skills refresh        Scan installed skills and write registry + assignments + digests
  shipwright skills status         Show indexed skills and warnings
  shipwright skills show <name>    Show one indexed skill
  shipwright skills assign         Use project-profile/planned stack and recommend skills by agent
  shipwright skills assign --json  Emit skill assignments as JSON
  shipwright skills providers      Show optional skill providers and embedded packs
  shipwright skills import autoskills
                                Import .agents/skills into .opencode/skills
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

func printSkillAssignmentsStatus(assignments *harness.SkillAssignmentSet) {
	fmt.Println("Shipwright — Skill Assignments")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println()
	fmt.Printf("Generated:      %s\n", assignments.GeneratedAt)
	fmt.Printf("Technologies:   %d\n", len(assignments.Technologies))
	fmt.Printf("Combos:         %d\n", len(assignments.Combos))
	fmt.Printf("Skills:         %d\n", len(assignments.Skills))
	fmt.Println()
	if len(assignments.Technologies) > 0 {
		fmt.Println("Detected technologies:")
		for _, tech := range assignments.Technologies {
			fmt.Printf("  - %-16s %s (%s)\n", tech.ID, tech.Name, strings.Join(tech.Evidence, ", "))
		}
		fmt.Println()
	}
	if len(assignments.Skills) > 0 {
		fmt.Println("Recommended skills:")
		for _, skill := range assignments.Skills {
			status := "missing"
			if skill.Installed {
				status = "installed"
			}
			fmt.Printf("  - %-28s %-9s agents=%s\n", skill.Name, status, strings.Join(skill.Agents, ","))
		}
	}
	if len(assignments.Warnings) > 0 {
		fmt.Println()
		fmt.Println("Warnings:")
		for _, warning := range assignments.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}
	fmt.Println()
	fmt.Println("Artifacts:")
	fmt.Println("  .harness/skill-assignments.json")
	fmt.Println("  .harness/skill-assignments.md")
}

func printSkillAssignmentsJSON(assignments *harness.SkillAssignmentSet) {
	data, err := json.MarshalIndent(assignments, "", "  ")
	if err != nil {
		Fail(fmt.Sprintf("cannot encode skill assignments: %s", err))
	}
	fmt.Println(string(data))
}

func printAutoSkillsImportJSON(result *harness.AutoSkillsImportResult) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		Fail(fmt.Sprintf("error serializando autoskills import: %s", err))
	}
	fmt.Println(string(data))
}

func printSkillProviders() {
	fmt.Println("Shipwright — Skill Providers")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println()
	fmt.Println("Embedded skill packs:")
	packs := harness.AllSkillPacks()
	if len(packs) == 0 {
		fmt.Println("  none")
	} else {
		for _, pack := range packs {
			fmt.Printf("  - %s (%d skills) — %s\n", pack.ID, len(pack.Skills), pack.Name)
		}
	}
	fmt.Println()
	fmt.Println("Optional providers:")
	if harness.AutoSkillsAvailable() {
		fmt.Println("  - autoskills: available at .agents/skills")
		fmt.Println("    import: shipwright skills import autoskills")
	} else {
		fmt.Println("  - autoskills: not detected (.agents/skills not found)")
	}
}
