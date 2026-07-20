package harness

import (
	"strings"
	"testing"
)

func TestAgentSkillsLoadFromProjectTemplates(t *testing.T) {
	skills := AllAgentSkills()
	if len(skills) != 7 {
		t.Fatalf("skills = %d, want 7", len(skills))
	}

	po := GetAgentSkill("product-owner")
	if po == nil {
		t.Fatal("product-owner skill not loaded")
	}
	if !strings.Contains(po.Content, "name: product-owner") {
		t.Fatalf("product-owner template content was not loaded")
	}
	if !strings.Contains(AgentCommonProtocol, "Shipwright Agent — Common Protocol") {
		t.Fatalf("shared agent common template was not loaded")
	}
}

func TestCuratedSkillTemplatesLoaded(t *testing.T) {
	frontend := GetCuratedSkill("frontend-design")
	if frontend == nil {
		t.Fatal("frontend-design curated skill not loaded")
	}
	if !strings.Contains(frontend.Content, "name: frontend-design") {
		t.Fatalf("frontend-design template content was not loaded")
	}
	reverse := GetCuratedSkill("existing-web-to-openpencil")
	if reverse == nil {
		t.Fatal("existing-web-to-openpencil curated skill not loaded")
	}
	if !strings.Contains(reverse.Content, "name: existing-web-to-openpencil") || !strings.Contains(reverse.Content, "Astro-specific guidance") {
		t.Fatalf("existing-web-to-openpencil template content was not loaded")
	}
	for _, required := range []string{
		".harness/artifacts/design/route-inventory.md",
		".harness/artifacts/design/fidelity-report.md",
		"Do not claim a faithful baseline from code inspection alone",
		"If any requested route/view was excluded without user approval, status is `fail`",
		"Section matrix per route/viewport",
		"If OpenPencil exports were not manually or visually inspected against source screenshots, status cannot be `pass`",
	} {
		if !strings.Contains(reverse.Content, required) {
			t.Fatalf("existing-web-to-openpencil missing fidelity guard %q", required)
		}
	}
	for _, skillName := range []string{"canvas-generate-design", "openpencil-generate-design", "design-code-component-map"} {
		skill := GetCuratedSkill(skillName)
		if skill == nil {
			t.Fatalf("%s curated skill not loaded", skillName)
		}
		if !strings.Contains(skill.Content, "name: "+skillName) {
			t.Fatalf("%s template content was not loaded", skillName)
		}
	}
	openpencil := GetCuratedSkill("openpencil-generate-design")
	for _, required := range []string{
		"## Save protocol",
		".harness/artifacts/design/openpencil/app.pen",
		".harness/artifacts/design/openpencil/save-status.md",
		"Do not declare `.pen` persistence unless save verification passed",
		"Inspect the current page/canvas before drawing",
		"read back page tree/node bounds",
		"primitive-only result must be documented as low-fidelity",
	} {
		if !strings.Contains(openpencil.Content, required) {
			t.Fatalf("openpencil-generate-design missing save guard %q", required)
		}
	}
	canvas := GetCuratedSkill("canvas-generate-design")
	for _, required := range []string{
		"Figma-inspired canvas discipline",
		"Component-first",
		"Token-first",
		"Readback required",
		"Primitive-only approximations of component-rich UI mean fail",
	} {
		if !strings.Contains(canvas.Content, required) {
			t.Fatalf("canvas-generate-design missing canvas discipline %q", required)
		}
	}
	designCodeMap := GetCuratedSkill("design-code-component-map")
	for _, required := range []string{
		"Design ↔ Code Component Map",
		".harness/artifacts/design/code-component-map.md",
		"DCC-001",
		"missing-code",
		"Official Figma Code Connect mode",
	} {
		if !strings.Contains(designCodeMap.Content, required) {
			t.Fatalf("design-code-component-map missing mapping rule %q", required)
		}
	}
	if len(AllCuratedSkills()) < 10 {
		t.Fatalf("expected curated lifecycle skills, got %d", len(AllCuratedSkills()))
	}
}

func TestSkillPackTemplatesLoaded(t *testing.T) {
	packs := AllSkillPacks()
	if len(packs) == 0 {
		t.Fatal("expected embedded skill packs")
	}
	found := false
	for _, pack := range packs {
		if pack.ID == "frontend-ui-quality" {
			found = true
			if len(pack.Skills) < 10 {
				t.Fatalf("frontend-ui-quality should include UI/UX skills: %+v", pack.Skills)
			}
			for _, skillName := range []string{"existing-web-to-openpencil", "canvas-generate-design", "openpencil-generate-design", "design-code-component-map"} {
				if !stringSliceHas(pack.Skills, skillName) {
					t.Fatalf("frontend-ui-quality should include %s: %+v", skillName, pack.Skills)
				}
			}
			if !stringSliceHas(pack.When.Stacks, "Astro") {
				t.Fatalf("frontend-ui-quality should match Astro: %+v", pack.When.Stacks)
			}
		}
	}
	if !found {
		t.Fatalf("frontend-ui-quality pack not found: %+v", packs)
	}
}

func stringSliceHas(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
