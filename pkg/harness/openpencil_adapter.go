package harness

import (
	"fmt"
	"strings"
)

type OpenPencilDesignAdapter struct{}

func NewOpenPencilDesignAdapter() *OpenPencilDesignAdapter {
	return &OpenPencilDesignAdapter{}
}

func (o *OpenPencilDesignAdapter) AdapterName() string {
	return DesignModeOpenPencil
}

func (o *OpenPencilDesignAdapter) StartDesign(state *State, request string) (*DesignResult, error) {
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

	if err := ensureDir(DesignOpenPencilDir); err != nil {
		return nil, fmt.Errorf("cannot create openpencil dir: %w", err)
	}
	if err := ensureDir(DesignExportsDir); err != nil {
		return nil, fmt.Errorf("cannot create exports dir: %w", err)
	}

	task := generateDesignTask(state, request)
	if err := WriteFile(DesignTaskFile, task); err != nil {
		return nil, fmt.Errorf("cannot write design-task.md: %w", err)
	}
	files = append(files, DesignTaskFile)

	penFile := "design/openpencil/app.pen"

	if err := SaveDesignState(DesignModeOpenPencil, false); err != nil {
		return nil, fmt.Errorf("cannot save design state: %w", err)
	}

	return &DesignResult{
		Adapter:      DesignModeOpenPencil,
		Mode:         DesignModeOpenPencil,
		FilesCreated: files,
		PenFile:      penFile,
		TaskFile:     DesignTaskFile,
		Message:      "OpenPencil design task created. AI agent should read design/openpencil/design-task.md and use open-pencil_* MCP tools to create the .pen file.",
	}, nil
}

func (o *OpenPencilDesignAdapter) Status() (*DesignStatus, error) {
	return &DesignStatus{
		Adapter:         DesignModeOpenPencil,
		Mode:            DesignModeOpenPencil,
		Available:       true,
		PenFile:         "design/openpencil/app.pen",
		HasBrief:        ArtifactExists("design/ux-brief.md"),
		HasFlows:        ArtifactExists("design/user-flows.md"),
		HasDecisions:    ArtifactExists("design/design-decisions.md"),
		HasPrototype:    ArtifactExists("design/prototype.md"),
		HasWireframes:   ArtifactExists("design/wireframes.md"),
		HasTaskFile:     ArtifactExists(DesignTaskFile),
		HasResponsiveQA: ArtifactExists("design/responsive-qa.md"),
	}, nil
}

func generateDesignTask(state *State, request string) string {
	var sb strings.Builder

	sb.WriteString("# OpenPencil Design Task\n\n")
	sb.WriteString(fmt.Sprintf("**Project:** %s\n", state.ProjectName))
	sb.WriteString(fmt.Sprintf("**Request:** %s\n\n", request))

	sb.WriteString("## Objective\n\n")
	sb.WriteString("Use the OpenPencil MCP tools to create a visual design for this project.\n\n")

	sb.WriteString("## MCP-first validation\n\n")
	sb.WriteString("Shipwright CLI detection can report `installed_no_active_canvas` even when OpenPencil desktop is open, because only the MCP client can verify the live editor/canvas. Do **not** treat that status as failure.\n\n")
	sb.WriteString("- First try the OpenCode MCP tools from the `open-pencil` server.\n")
	sb.WriteString("- Expected OpenCode tool pattern: `open-pencil_*` because MCP tools are prefixed with the server name.\n")
	sb.WriteString("- If another MCP server named `pencil` is also connected, do **not** use it for Shipwright OpenPencil work; it can be bound to another app host.\n")
	sb.WriteString("- Only fall back to doc-only design if no `open-pencil_*` MCP tool is visible or the editor-state check fails.\n\n")

	sb.WriteString("## Steps\n\n")
	sb.WriteString("1. Call `open-pencil_get_editor_state` to verify OpenPencil is connected and an editor/canvas is reachable.\n")
	sb.WriteString("2. If no active canvas exists but the MCP responds, create or open one with `open-pencil_batch_design` or use the existing canvas.\n")
	sb.WriteString("3. Read `design/ux-brief.md` for product context and design requirements.\n")
	sb.WriteString("4. Read `design/user-flows.md` for the flows to design.\n")
	sb.WriteString("5. Create the design file at `design/openpencil/app.pen`.\n")
	sb.WriteString("6. Create responsive frames for every key screen: mobile 390×844, tablet 768×1024, desktop 1440×1024.\n")
	sb.WriteString("7. Design mobile-first, then adapt tablet and desktop. Do not simply scale the same layout.\n")
	sb.WriteString("8. Export wireframes to `design/openpencil/exports/` using `open-pencil_export_nodes`.\n")
	sb.WriteString("9. Take screenshots with `open-pencil_get_screenshot` and inspect them for overflow, clipping, hidden text, tiny tap targets, and horizontal scroll.\n")
	sb.WriteString("10. Create `design/responsive-qa.md` with the responsive/accessibility checks and findings.\n")
	sb.WriteString("11. Create `design/prototype.md` describing the visual design and how it maps to user flows.\n")
	sb.WriteString("12. Run `shipwright design status` to verify all artifacts are in place.\n\n")

	sb.WriteString("## Responsive Layout Contract\n\n")
	sb.WriteString("- Use an 8px spacing grid and keep outer safe margins: 16px mobile, 24px tablet, 32px desktop.\n")
	sb.WriteString("- No component may extend outside its frame/canvas.\n")
	sb.WriteString("- Avoid fixed-width containers that exceed the frame width.\n")
	sb.WriteString("- Primary actions must remain visible without horizontal scrolling.\n")
	sb.WriteString("- Interactive targets must be at least 44×44px.\n")
	sb.WriteString("- Body text should be at least 16px with readable line-height.\n")
	sb.WriteString("- Color contrast must target WCAG AA: 4.5:1 normal text, 3:1 large text/UI components.\n")
	sb.WriteString("- Every screen must include loading, empty, error, and success states where relevant.\n\n")

	sb.WriteString("## Rules\n\n")
	sb.WriteString("- **NEVER** read `.pen` files directly with filesystem tools.\n")
	sb.WriteString("- **ONLY** manipulate `.pen` files via OpenPencil MCP tools from the `open-pencil` server (`open-pencil_*`).\n")
	sb.WriteString("- **DO NOT** mark design complete if any screenshot shows overflowing components, clipped text, or content outside the canvas.\n")
	sb.WriteString("- The `.pen` file is NOT considered approved just because it exists.\n")
	sb.WriteString("- Human approval via `shipwright approve ux-design` is still **mandatory**.\n\n")

	sb.WriteString("## Required artifacts after completion\n\n")
	sb.WriteString("- `design/openpencil/app.pen` — the design file\n")
	sb.WriteString("- `design/openpencil/exports/` — exported wireframes/prototypes\n")
	sb.WriteString("- `design/responsive-qa.md` — responsive/accessibility validation\n")
	sb.WriteString("- `design/prototype.md` — description of the visual design\n\n")

	sb.WriteString("## After completion\n\n")
	sb.WriteString("Run: `shipwright next` to advance to UX_APPROVAL.\n")
	sb.WriteString("Then the user must approve: `shipwright approve ux-design`\n")

	return sb.String()
}

func ensureDir(path string) error {
	return WriteFile(path+"/.gitkeep", "")
}
