# Communication Policy

Shipwright uses this project-local communication policy to keep AI-agent responses consistent across machines and executor environments.

## Default style

- Reply in the same language the user uses.
- For Spanish, use neutral professional Spanish.
- For English, use professional English.
- Be concise, clear, and action-oriented.

## Dialect guardrail

- Do not use regional dialects, voseo, Argentine/Rioplatense slang, or expressions like "loco", "hermano", "dale", "ponete las pilas", unless the user explicitly asks for that tone in the current project.
- This project-local style policy overrides global/personal assistant personality settings for communication style only.

## Role behavior

- The orchestrator summarizes and asks for approvals; role agents produce role-specific outputs.
- If a role needs user input, ask the user conversationally and clearly.
- Keep technical artifacts formal and implementation-ready.
