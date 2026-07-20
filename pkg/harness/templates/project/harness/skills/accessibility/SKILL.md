---
name: accessibility
description: >
  Accessibility and inclusive design review skill for WCAG-minded UI decisions.
  Trigger: Use when designing, reviewing, or implementing screens, forms, navigation, dialogs, tables, or interactive components.
license: Apache-2.0
metadata:
  author: shipwright
  version: "1.0"
---

## When to Use

Use this skill for every user-facing UI artifact before approving UX or frontend implementation.

## Critical Patterns

1. Keyboard first: every interactive element must be reachable and operable by keyboard.
2. Visible focus is mandatory; never remove outlines without an accessible replacement.
3. Text and controls need WCAG AA contrast; do not rely on color alone.
4. Touch targets should be at least 44×44 CSS pixels for primary mobile interactions.
5. Dialogs, sheets, drawers, menus, and popovers require focus management and accessible names.
6. Forms require explicit labels, error messages, and field-level recovery guidance.

## Review Checklist

- Semantic landmarks: header, nav, main, aside, footer where applicable.
- Heading order does not skip levels for visual styling.
- Images/icons have useful alt text or are explicitly decorative.
- Tables include headers and responsive alternatives when too dense for mobile.
- Motion respects reduced-motion preferences.

## Evidence

Document findings and fixes in `.harness/artifacts/design/responsive-qa.md` or `.harness/artifacts/reports/qa-report.md`.
