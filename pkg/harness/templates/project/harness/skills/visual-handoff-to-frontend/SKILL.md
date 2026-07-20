---
name: visual-handoff-to-frontend
description: >
  Visual handoff skill for turning approved UX/UI designs into frontend-ready implementation guidance.
  Trigger: Use after UX approval and before frontend implementation tasks begin.
license: Apache-2.0
metadata:
  author: shipwright
  version: "1.0"
---

## When to Use

Use this skill when moving from approved design to frontend tasks, mocks, components, and implementation constraints.

## Critical Patterns

1. Handoff must include screens, components, states, tokens, data needs, and responsive behavior.
2. Frontend must preserve mock mode and real API mode when Shipwright contract-first rules apply.
3. Identify reusable components before implementation starts.
4. Include accessibility requirements in the task, not as a later review surprise.
5. Every UI task should reference its source design artifact.

## Handoff Checklist

- Screens and routes.
- Component inventory.
- Props/data needed per component.
- Empty/loading/error states.
- Responsive behavior per breakpoint.
- Accessibility notes.
- Open questions or risks.

## Output

Update `.harness/artifacts/backlog/frontend-tasks.md`, `.harness/artifacts/design/design-decisions.md`, and relevant SDD tasks.
