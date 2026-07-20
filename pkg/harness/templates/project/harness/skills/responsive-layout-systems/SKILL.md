---
name: responsive-layout-systems
description: >
  Responsive layout systems skill for mobile, tablet, and desktop UI design.
  Trigger: Use when creating layouts, grids, dashboards, tables, forms, or responsive prototypes.
license: Apache-2.0
metadata:
  author: shipwright
  version: "1.0"
---

## When to Use

Use this skill whenever a screen must work across mobile, tablet, and desktop breakpoints.

## Critical Patterns

1. Design mobile first, then enhance for tablet and desktop.
2. Use fluid containers and predictable max widths; avoid fixed pixel layouts for content regions.
3. Define how navigation transforms: bottom tabs, drawer, sidebar, top nav, or command menu.
4. Dense tables need mobile alternatives: cards, stacked rows, column priority, or drill-down.
5. Prevent overflow: no clipped cards, hidden CTAs, unreadable labels, or forced horizontal scrolling unless intentionally designed.

## Required Breakpoints

- Mobile: 390×844
- Tablet: 768×1024
- Desktop: 1440×1024

## QA Rules

For each key screen, verify: no overflow, readable typography, usable touch targets, preserved primary action, and stable navigation.
