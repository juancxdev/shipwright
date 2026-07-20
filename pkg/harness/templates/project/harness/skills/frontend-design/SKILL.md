---
name: frontend-design
description: >
  Product-aware frontend UI design skill for turning requirements into clear, usable, responsive interfaces.
  Trigger: Use when designing web app screens, dashboards, forms, navigation, empty states, or visual prototypes.
license: Apache-2.0
metadata:
  author: shipwright
  version: "1.0"
---

## When to Use

Use this skill when creating or reviewing UI concepts, wireframes, prototypes, or screen specifications for frontend applications.

## Critical Patterns

1. Start from user goals, not components. Every screen must state the primary user action.
2. Design complete states: loading, empty, error, success, disabled, validation, and permission-denied.
3. Prioritize hierarchy: one primary action per view, clear secondary actions, no competing CTAs.
4. Prefer simple layouts that scale: header, content region, contextual aside, and predictable navigation.
5. Never declare a design complete without mobile, tablet, and desktop behavior.

## UI Quality Checklist

- Screen purpose is obvious within 5 seconds.
- Primary action is visible without hunting.
- Forms have labels, help text, validation, and recovery paths.
- Tables/lists define sorting, filtering, empty state, and pagination/infinite-scroll behavior.
- Dashboard cards show useful comparisons, not decorative numbers.

## Handoff Output

Record final design decisions in `.harness/artifacts/design/prototype.md`, `.harness/artifacts/design/design-decisions.md`, and `.harness/artifacts/design/responsive-qa.md`.
