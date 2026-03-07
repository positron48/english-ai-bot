---
name: setup-agent-swarm
description: Designs and creates a full Cursor agent swarm for the user's task. Uses reference.md in this skill folder as the single source of truth for .cursor/ structure, subagents, commands, rules, skills, and mcp_task. Use when the user wants to set up a swarm of agents, create workflows with subagents, or configure Cursor for multi-role tasks (e.g. code review, deployment, refactoring). Produces a plan for approval, then creates agents, commands, rules, and skills.
---

# Agent swarm setup

Create and configure an agent swarm (subagents, commands, rules, skills) for the user's task. The single source of truth for structure and formats is **reference.md** in this skill folder.

---

## When to use

- User asks to “set up an agent swarm for …”, “create an agent workflow”, or “configure Cursor so that …” with multiple roles or steps.
- A process is described (code review, deploy, refactor, docs) that should be automated by a swarm.
- You need to add subagents, commands, rules, and skills and wire them via mcp_task.

---

## Source of truth

Before creating files, read **reference.md** (in this skill folder) and follow:

- Folder structure under `.cursor/` (agents, commands, rules, skills, templates, scripts, config.json, hooks.json).
- Subagent format: `.cursor/agents/<role>.md`; invocation only via **mcp_task** with `subagent_type` = role name without `.md`.
- Command format: `.cursor/commands/<name>.md`; in steps explicitly call `mcp_task(subagent_type="…", prompt="…", description="…)`, do not perform the role as the orchestrator.
- Built-in Cursor commands: `/create-subagent`, `/create-rule`, `/create-skill`, `/migrate-to-skills`.
- mcp_task parameters: `subagent_type`, `prompt`, `description`, optionally `resume`.

Full checklist and formats are in [reference.md](reference.md).

---

## Phase 1 — Plan (required before creation)

1. **Clarify the task**
   - Goal: which process or workflow should the swarm perform?
   - If unclear, ask 1–2 short questions. In one sentence: “The swarm will …”.

2. **Design the swarm** per reference.md:
   - **Subagents:** roles (planner, worker, reviewer, test-runner, documenter, etc.). Each role = `.cursor/agents/<role>.md`, invoked via `mcp_task(subagent_type="<role>", …)`.
   - **Commands:** slash commands (e.g. `/full-feature`, `/quick-fix`) that orchestrate steps. Each = `.cursor/commands/<name>.md`; in steps — mcp_task calls, not performing the role “as self”.
   - **Rules:** which rules to apply always or contextually (.cursor/rules/*.mdc).
   - **Skills:** reusable skills for subagents/orchestrator (.cursor/skills/<name>/SKILL.md or .agents/skills/).
   - **Interaction:** which command calls whom and in what order; short text flow (e.g. `/foo → planner → worker → test-runner`).

3. **Present the plan for approval**
   - One block with a heading like “## Plan: Agent swarm for [goal]”.
   - Include: short summary; subagents (name, purpose, skills); commands (name, steps with mcp_task); rules; skills; interaction diagram.
   - End with: “Approve the plan (reply ‘approve’ / ‘ok’ / ‘create’), or describe changes.”
   - **Do not create** any files under .cursor/ until the user explicitly approves.

4. **Incorporate feedback**
   - If the user requests changes — update the plan and show it again with the same approval prompt. Repeat until approved.

---

## Phase 2 — Creation (only after explicit approval)

1. **Confirm** — briefly state that you are creating the swarm per the approved plan.

2. **Create in dependency order**
   - **Skills** (if any): use **/create-skill** for each.
   - **Rules:** **/create-rule** for each.
   - **Subagents:** **/create-subagent** for each role (role description, what it does and does not do, which skills).
   - **Commands:** create manually `.cursor/commands/<name>.md` with title, usage, steps (for subagent calls — explicitly “call mcp_task with subagent_type=…, prompt=…, description=…”), result, and notes.
   - **config.json** (if in the plan): add or update `.cursor/config.json` as needed.

3. **Do not add** unless requested: hooks.json, templates, scripts — only if the user or plan specifies them.

4. **Summary** — list what was created (agents, commands, rules, skills) and remind how to run (e.g. “Use /command-name; subagents are invoked via mcp_task.”).

---

## Principles

- **Only via mcp_task:** in commands, subagent roles are invoked exclusively by calling mcp_task; the orchestrator does not “play” the subagent role.
- **Single approval gate:** no files until the user explicitly approves the plan.
- **Follow the reference:** structure and formats are defined in reference.md in this skill folder.
