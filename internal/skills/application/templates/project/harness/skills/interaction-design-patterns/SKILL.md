---
name: interaction-design-patterns
description: >
  Interaction design skill for user flows, state transitions, feedback, dialogs, and complex UI behavior.
  Trigger: Use when designing multi-step flows, forms, dashboards, CRUD screens, confirmations, or error recovery.
license: Apache-2.0
metadata:
  author: shipwright
  version: "1.0"
---

## When to Use

Use this skill when screens are not static and users need to complete tasks reliably.

## Critical Patterns

1. Every user action needs feedback: immediate, understandable, and recoverable.
2. Destructive actions require confirmation, impact explanation, and safe cancellation.
3. Multi-step flows must show progress, allow review, and preserve user input.
4. Errors should explain what happened, why it matters, and how to recover.
5. Avoid modal abuse; use inline editing, drawers, or dedicated pages when context matters.

## Required States

For each key flow define: idle, loading, optimistic/pending, success, validation error, system error, empty, and permission denied.

## Handoff

Record interaction decisions in `.harness/artifacts/design/user-flows.md` and implementation notes in backlog tasks.
