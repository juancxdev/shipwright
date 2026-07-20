---
name: design-system-tokens
description: >
  Design system token skill for consistent color, typography, spacing, radius, shadows, and component decisions.
  Trigger: Use when defining visual style, design systems, UI kits, or frontend handoff tokens.
license: Apache-2.0
metadata:
  author: shipwright
  version: "1.0"
---

## When to Use

Use this skill when a design needs consistent visual language or when UI decisions must be handed off to frontend engineers.

## Critical Patterns

1. Define tokens before screens become inconsistent.
2. Keep palettes small: semantic colors beat arbitrary color names.
3. Typography must include scale, weight, line-height, and usage rules.
4. Spacing should use a consistent rhythm; avoid one-off values.
5. Components need variants, states, and accessibility behavior.

## Minimum Token Set

- Color: background, foreground, muted, primary, secondary, destructive, border, focus.
- Typography: display, heading, body, caption, mono if needed.
- Spacing: 4/8-based scale or explicitly justified alternative.
- Radius: small, medium, large, full.
- Elevation: none, subtle, raised, overlay.

## Handoff

Write token decisions in `.harness/artifacts/design/design-decisions.md` and reference them from frontend tasks.
