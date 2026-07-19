package cmd

import (
	"fmt"
	"os"
)

const UsageText = `Shipwright — Agentic Software Delivery Harness

Usage:
  shipwright init [--ai opencode] [--reasoning-model <model>] [--fast-model <model>]
                                      Initialize harness structure in current directory
  shipwright start "<request>"       Start a new delivery cycle with a user request
  shipwright status                        Show current phase, approvals, and next action
  shipwright next                          Advance to next phase if gates are met
  shipwright run                     Auto-scaffold + advance through all phases (stops at gates)
  shipwright approve <gate>                Approve a gate (scope|ux-design|technical-plan|tech-lead|final-acceptance)
  shipwright request-change ["<reason>"]
                                      Request a change (triggers change request flow)
  shipwright scaffold                Generate placeholder artifacts for current phase
  shipwright generate <artifact>     Generate a single artifact from template
  shipwright generate --list         List all scaffoldable artifacts
  shipwright agents <subcommand>           Agent management (list|show|active)
  shipwright contract <subcommand>   Contract management (validate|generate-tasks|check-mocks|check-compliance|show)
  shipwright memory <subcommand>     Memory management (status|enable|disable|flush|mark-synced)
  shipwright integrations <subcommand>     Integrations (status|enable|disable|detect)
  shipwright config <subcommand>     Portable config (show|init|env|validate)
  shipwright executor <subcommand>         Executor adapters (list|status|generate)
  shipwright doctor [--json] [--fix]       Diagnose/fix config, integrations and fallbacks
  shipwright design <subcommand>     Design management (start|status)
  shipwright review <subcommand>     Review evidence (start|status)
  shipwright skills <subcommand>     Skill registry/digests (refresh|status|show|digest)
  shipwright tdd <subcommand>        TDD policy/evidence gate (refresh|status|policy)

Config:
  show                Show effective portable config (defaults + file + platform + env)
  init                Create .harness/config.json for older projects
  env                 List supported environment overrides
  validate            Validate .harness/config.json semantically
  validate --json     Emit machine-readable validation and exit 2 on errors

Executor:
  init                Defaults to OpenCode executor; --executor opencode remains supported
  init --ai opencode  Explicitly select OpenCode as AI executor
  init --executor opencode
                      Backward-compatible alias for --ai opencode
  list                Show available executor adapters
  status [name]       Show generated executor files and missing pieces
  generate <name>     Generate executor bootstrap (generic|opencode)
  generate opencode --reasoning-model <model> --fast-model <model>
                      Set OpenCode role model defaults and regenerate
  generate opencode --agent-model role=model
                      Override one OpenCode role model and regenerate

Doctor:
  doctor              Show actionable diagnostics for platform/config/integrations
  doctor --json       Emit machine-readable diagnostics and exit 2 on blocking errors
  doctor --fix        Safely create, normalize, or back up/recreate portable config

Gates:
  scope               Approve functional scope (valid in SCOPE_REVIEW)
  ux-design           Approve UX design (valid in UX_APPROVAL)
  technical-plan      Approve technical plan (valid in BACKLOG_READY)
  tech-lead           Approve tech lead review (valid in TECH_LEAD_REVIEW)
  final-acceptance    Accept final delivery (valid in USER_ACCEPTANCE)

Agents:
  list                Show all 7 agents with purpose and phases
  show <name>         Show full skill definition (frontmatter + steps + rules)
  active              Show active agent for current phase + steps + done criteria
  run <name>          Output SKILL.md for AI execution (ready to paste into agent prompt)

Scaffold:
  scaffold            Generate all placeholder artifacts for the current phase
  generate <path>     Generate a single artifact (use --list to see options)
  run                 Walk the full lifecycle: scaffold + advance, stop at approval gates

Contract:
  validate            Validate contracts/openapi.yaml structure (endpoints, schemas, errors)
  generate-tasks      Generate backlog/frontend-tasks.md + backend-tasks.md from contract
  check-mocks         Verify frontend mock mode compliance (mocks mandatory, preserved)
  check-compliance    Verify backend API matches contract (endpoints, schemas, errors)
  show                Display parsed contract summary (endpoints, schemas, auth)

Review:
  start               Generate QA/security/contract reports and review checklist
  status              Validate review evidence and finding severity gates

Skills:
  refresh             Scan project skill sources and update .harness/skill-registry.*
  status              Show indexed skills and warnings
  show <name>         Show one indexed skill
  digest [agent]      Show compact skill rules for all agents or one agent

TDD:
  refresh             Regenerate .harness/tdd-policy.* from project calibration
  status              Show strict/suggested/none mode and evidence status
  policy              Print the current TDD policy markdown

Aliases:
  Documentation may still refer to 'harness'; the distributed binary is 'shipwright'.
`

func PrintUsage() {
	fmt.Print(UsageText)
}

func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
}

func PrintSuccess(msg string) {
	fmt.Printf("✓ %s\n", msg)
}

func PrintInfo(msg string) {
	fmt.Printf("  %s\n", msg)
}

func Fail(msg string) {
	PrintError(msg)
	os.Exit(1)
}

func Exit(code int) {
	os.Exit(code)
}

func EnsureHarness() {
	if !harnessInitialized() {
		Fail("harness no inicializado. Ejecutá 'shipwright init' primero.")
	}
}

func harnessInitialized() bool {
	info, err := os.Stat(".harness/state.json")
	return err == nil && !info.IsDir()
}
