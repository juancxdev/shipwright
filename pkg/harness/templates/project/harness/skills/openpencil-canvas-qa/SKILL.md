---
name: openpencil-canvas-qa
description: >
  OpenPencil canvas QA skill for validating generated designs, frames, exports, clipping, overflow, and responsive behavior.
  Trigger: Use when OpenPencil MCP is enabled or when converting doc-only wireframes into OpenPencil prototypes.
license: Apache-2.0
metadata:
  author: shipwright
  version: "1.0"
---

## When to Use

Use this skill when working with OpenPencil, `.pen` files, exported screenshots, or design frames.

## Critical Patterns

1. Do not assume CLI detection proves the desktop canvas is active; verify with the actual MCP tool when available.
2. Always inspect generated frames before declaring UX ready.
3. Create separate frames for mobile, tablet, and desktop.
4. No component may extend outside its frame/canvas.
5. Export screenshots and check for clipping, overlap, unreadable text, and missing states.

## Validation Order

1. Try OpenCode MCP tools for `open-pencil`.
2. Read editor/canvas state.
3. Generate or update frames.
4. Export screenshots.
5. Record QA evidence in `.harness/artifacts/design/responsive-qa.md`.

## Fallback

If OpenPencil is unavailable, use doc-only wireframes but clearly mark the fallback and keep responsive QA mandatory.
