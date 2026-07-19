# Shipwright Release Hardening Checklist

Target: stabilize the harness before moving from v0.8.x hardening to v0.9.0.

## CLI smoke

- [ ] `harness help` shows `config`, `doctor`, `integrations`, `review`, `contract`.
- [ ] `shipwright init` creates `.harness/config.json` and `.harness/integrations.json`.
- [ ] `shipwright config validate` passes on a fresh project.
- [ ] `shipwright doctor --json` returns `summary`, `engram_health`, `openpencil_health`.
- [ ] `shipwright doctor --fix` creates config when missing.
- [ ] `shipwright doctor --fix` backs up corrupt config before recreating.

## Cross-platform

- [ ] macOS: detects `/Applications/OpenPencil.app` candidates.
- [ ] Windows: supports `ENGRAM_BINARY` and `OPENPENCIL_MCP_SERVER` paths.
- [ ] Linux: supports `/opt/OpenPencil` and `$HOME/.local/share/OpenPencil` candidates.
- [ ] CI: can run with no Engram/OpenPencil installed and still use fallbacks.

## Safety

- [ ] Health checks have timeout.
- [ ] OpenPencil MCP is not executed during doctor.
- [ ] Config fixes do not delete custom paths.
- [ ] Corrupt config is backed up before replacement.
- [ ] Optional integration failures are warnings, not blocking errors.

## Evidence gates

- [ ] Review phase still blocks on critical/high findings.
- [ ] Medium findings require explicit decision.
- [ ] QA/security/contract reports exist before user acceptance.
- [ ] Contract-first FE/BE gates still validate mocks and backend compliance.

## Tests

- [ ] `go test ./...` passes.
- [ ] Config validation tests pass.
- [ ] Doctor/fix tests pass.
- [ ] Integration detection tests pass.
- [ ] State machine/gate hardening tests pass.
- [ ] CLI smoke test passes.

## Known future work

- [ ] Real OpenPencil MCP protocol handshake when stable non-interactive API exists.
- [ ] Optional `doctor --fix --aggressive` for explicitly clearing invalid custom paths.
- [ ] Optional `doctor --ci` profile with stricter CI semantics.
