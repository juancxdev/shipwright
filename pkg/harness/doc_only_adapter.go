package harness

import (
	"fmt"
	"strings"
)

type DocOnlyDesignFallback struct{}

func NewDocOnlyDesignFallback() *DocOnlyDesignFallback {
	return &DocOnlyDesignFallback{}
}

func (d *DocOnlyDesignFallback) AdapterName() string {
	return DesignModeDocOnly
}

func (d *DocOnlyDesignFallback) StartDesign(state *State, request string) (*DesignResult, error) {
	var files []string

	brief := generateUXBrief(state, request)
	if err := WriteFile("design/ux-brief.md", brief); err != nil {
		return nil, fmt.Errorf("cannot write ux-brief.md: %w", err)
	}
	files = append(files, "design/ux-brief.md")

	flows := generateUserFlows(state)
	if err := WriteFile("design/user-flows.md", flows); err != nil {
		return nil, fmt.Errorf("cannot write user-flows.md: %w", err)
	}
	files = append(files, "design/user-flows.md")

	decisions := generateDesignDecisions(state)
	if err := WriteFile("design/design-decisions.md", decisions); err != nil {
		return nil, fmt.Errorf("cannot write design-decisions.md: %w", err)
	}
	files = append(files, "design/design-decisions.md")

	wireframes := generateWireframesDoc(state)
	if err := WriteFile("design/wireframes.md", wireframes); err != nil {
		return nil, fmt.Errorf("cannot write wireframes.md: %w", err)
	}
	files = append(files, "design/wireframes.md")

	prototype := generatePrototypeDoc(state)
	if err := WriteFile("design/prototype.md", prototype); err != nil {
		return nil, fmt.Errorf("cannot write prototype.md: %w", err)
	}
	files = append(files, "design/prototype.md")

	responsiveQA := generateResponsiveQADoc(state)
	if err := WriteFile("design/responsive-qa.md", responsiveQA); err != nil {
		return nil, fmt.Errorf("cannot write responsive-qa.md: %w", err)
	}
	files = append(files, "design/responsive-qa.md")

	if err := SaveDesignState(DesignModeDocOnly, true); err != nil {
		return nil, fmt.Errorf("cannot save design state: %w", err)
	}

	return &DesignResult{
		Adapter:      DesignModeDocOnly,
		Mode:         DesignModeDocOnly,
		FilesCreated: files,
		Message:      "OpenPencil unavailable: design generated in doc-only mode.",
		FallbackUsed: true,
	}, nil
}

func (d *DocOnlyDesignFallback) Status() (*DesignStatus, error) {
	mode, _, _ := LoadDesignState()

	return &DesignStatus{
		Adapter:         DesignModeDocOnly,
		Mode:            mode,
		Available:       false,
		HasBrief:        ArtifactExists("design/ux-brief.md"),
		HasFlows:        ArtifactExists("design/user-flows.md"),
		HasDecisions:    ArtifactExists("design/design-decisions.md"),
		HasPrototype:    ArtifactExists("design/prototype.md"),
		HasWireframes:   ArtifactExists("design/wireframes.md"),
		HasTaskFile:     false,
		HasResponsiveQA: ArtifactExists("design/responsive-qa.md"),
	}, nil
}

func generateUXBrief(state *State, request string) string {
	var sb strings.Builder

	sb.WriteString("# UX Brief\n\n")
	sb.WriteString(fmt.Sprintf("**Project:** %s\n", state.ProjectName))
	sb.WriteString(fmt.Sprintf("**Request:** %s\n\n", request))

	sb.WriteString("## Product context\n\n")
	sb.WriteString("(Complete from product/context.md and product/scope.md)\n\n")

	sb.WriteString("## Target users\n\n")
	sb.WriteString("(Who will use this product? Define personas.)\n\n")

	sb.WriteString("## Key user goals\n\n")
	sb.WriteString("(What should users achieve with this product?)\n\n")

	sb.WriteString("## Design constraints\n\n")
	sb.WriteString("- Brand guidelines: (if any)\n")
	sb.WriteString("- Platform: (web, mobile, desktop)\n")
	sb.WriteString("- Accessibility: (WCAG level, etc.)\n\n")

	sb.WriteString("## Visual style\n\n")
	sb.WriteString("- Tone: (professional, playful, minimal, etc.)\n")
	sb.WriteString("- Colors: (primary, secondary, accent)\n")
	sb.WriteString("- Typography: (heading, body fonts)\n\n")

	sb.WriteString("## Key screens to design\n\n")
	sb.WriteString("(List the main screens/views needed)\n\n")

	return sb.String()
}

func generateUserFlows(state *State) string {
	var sb strings.Builder

	sb.WriteString("# User Flows\n\n")
	sb.WriteString(fmt.Sprintf("**Project:** %s\n\n", state.ProjectName))

	sb.WriteString("## Flow 1: Primary user journey\n\n")
	sb.WriteString("```\n[Entry point] -> [Step 1] -> [Step 2] -> [Goal]\n```\n\n")
	sb.WriteString("**Description:** (Describe the primary flow)\n\n")

	sb.WriteString("## Flow 2: Secondary flow\n\n")
	sb.WriteString("```\n[Entry point] -> [Step 1] -> [Decision] -> [Path A / Path B]\n```\n\n")
	sb.WriteString("**Description:** (Describe the secondary flow)\n\n")

	sb.WriteString("## Error flows\n\n")
	sb.WriteString("(What happens when things go wrong?)\n\n")

	return sb.String()
}

func generateDesignDecisions(state *State) string {
	var sb strings.Builder

	sb.WriteString("# Design Decisions\n\n")
	sb.WriteString(fmt.Sprintf("**Project:** %s\n\n", state.ProjectName))

	sb.WriteString("## Decision log\n\n")
	sb.WriteString("| # | Decision | Rationale | Date |\n")
	sb.WriteString("|---|----------|-----------|------|\n")
	sb.WriteString("| 1 | (decision) | (why) | (date) |\n\n")

	sb.WriteString("## Design principles\n\n")
	sb.WriteString("- (Add design principles for this project)\n\n")

	sb.WriteString("## Component inventory\n\n")
	sb.WriteString("(List reusable UI components needed)\n\n")

	return sb.String()
}

func generateWireframesDoc(state *State) string {
	var sb strings.Builder

	sb.WriteString("# Wireframes (Doc-Only Mode)\n\n")
	sb.WriteString(fmt.Sprintf("**Project:** %s\n\n", state.ProjectName))

	sb.WriteString("> **Note:** OpenPencil unavailable: design generated in doc-only mode.\n")
	sb.WriteString("> These are text-based wireframes. For visual design, enable OpenPencil:\n")
	sb.WriteString("> `shipwright integrations enable openpencil`\n\n")

	sb.WriteString("## Screen 1: (name)\n\n")
	sb.WriteString("```\n")
	sb.WriteString("+------------------------------------------+\n")
	sb.WriteString("|  [Header / Logo]              [Menu]    |\n")
	sb.WriteString("+------------------------------------------+\n")
	sb.WriteString("|                                          |\n")
	sb.WriteString("|  [Main content area]                     |\n")
	sb.WriteString("|                                          |\n")
	sb.WriteString("|  [Action button]                         |\n")
	sb.WriteString("|                                          |\n")
	sb.WriteString("+------------------------------------------+\n")
	sb.WriteString("```\n\n")
	sb.WriteString("**Description:** (Describe this screen)\n\n")

	sb.WriteString("## Screen 2: (name)\n\n")
	sb.WriteString("```\n")
	sb.WriteString("+------------------------------------------+\n")
	sb.WriteString("|  [Header]                               |\n")
	sb.WriteString("+------------------------------------------+\n")
	sb.WriteString("|  [Form / Input fields]                   |\n")
	sb.WriteString("|                                          |\n")
	sb.WriteString("|  [Submit] [Cancel]                       |\n")
	sb.WriteString("+------------------------------------------+\n")
	sb.WriteString("```\n\n")
	sb.WriteString("**Description:** (Describe this screen)\n\n")

	return sb.String()
}

func generatePrototypeDoc(state *State) string {
	var sb strings.Builder

	sb.WriteString("# Prototype Description (Doc-Only Mode)\n\n")
	sb.WriteString(fmt.Sprintf("**Project:** %s\n\n", state.ProjectName))

	sb.WriteString("> **Note:** OpenPencil unavailable: design generated in doc-only mode.\n")
	sb.WriteString("> This is a text-based prototype description.\n\n")

	sb.WriteString("## Screen flow\n\n")
	sb.WriteString("```\n")
	sb.WriteString("[Screen 1] --click--> [Screen 2] --submit--> [Screen 3: Success]\n")
	sb.WriteString("                              |\n")
	sb.WriteString("                              +--error--> [Screen 4: Error]\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Interaction notes\n\n")
	sb.WriteString("- (Describe key interactions)\n\n")

	sb.WriteString("## States\n\n")
	sb.WriteString("- **Loading:** (what the user sees while waiting)\n")
	sb.WriteString("- **Empty:** (what the user sees with no data)\n")
	sb.WriteString("- **Error:** (what the user sees when something fails)\n")
	sb.WriteString("- **Success:** (what the user sees on success)\n\n")

	return sb.String()
}

func generateResponsiveQADoc(state *State) string {
	var sb strings.Builder

	sb.WriteString("# Responsive & Accessibility QA\n\n")
	sb.WriteString(fmt.Sprintf("**Project:** %s\n\n", state.ProjectName))

	sb.WriteString("## Breakpoints checked\n\n")
	sb.WriteString("| Screen | Mobile 390x844 | Tablet 768x1024 | Desktop 1440x1024 | Notes |\n")
	sb.WriteString("|--------|----------------|-----------------|-------------------|-------|\n")
	sb.WriteString("| Screen 1 | Pending | Pending | Pending | Verify overflow, clipping, spacing |\n")
	sb.WriteString("| Screen 2 | Pending | Pending | Pending | Verify overflow, clipping, spacing |\n\n")

	sb.WriteString("## Layout checks\n\n")
	sb.WriteString("- [ ] No component extends outside its frame/canvas\n")
	sb.WriteString("- [ ] No horizontal scrolling is required\n")
	sb.WriteString("- [ ] Content uses safe margins: 16px mobile, 24px tablet, 32px desktop\n")
	sb.WriteString("- [ ] Layout adapts, it is not just scaled\n")
	sb.WriteString("- [ ] Primary action remains visible and reachable\n")
	sb.WriteString("- [ ] Empty/loading/error/success states are designed where relevant\n\n")

	sb.WriteString("## Accessibility checks\n\n")
	sb.WriteString("- [ ] Touch targets are at least 44x44px\n")
	sb.WriteString("- [ ] Body text is at least 16px and readable\n")
	sb.WriteString("- [ ] Contrast targets WCAG AA: 4.5:1 normal text, 3:1 large text/UI components\n")
	sb.WriteString("- [ ] Focus order and keyboard flow are logical\n")
	sb.WriteString("- [ ] Icon-only actions have text labels or accessible names\n\n")

	sb.WriteString("## Visual quality checks\n\n")
	sb.WriteString("- [ ] The design has a deliberate visual direction tied to the product context\n")
	sb.WriteString("- [ ] Typography, color, and spacing use consistent tokens\n")
	sb.WriteString("- [ ] Components are reusable and consistent across screens\n")
	sb.WriteString("- [ ] The design avoids generic template-looking UI unless intentionally justified\n\n")

	sb.WriteString("## Fixes applied\n\n")
	sb.WriteString("- Pending\n\n")

	return sb.String()
}
