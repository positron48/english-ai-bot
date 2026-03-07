# Reference: Agent swarm structure and formats

Single source of truth for creating and configuring agent swarms in Cursor. Subagents are invoked via the **mcp_task** tool.

---

## 1. Folder structure

Everything lives under `.cursor/` in the project root. Cursor also loads skills from **`.agents/skills/`** and **`~/.cursor/skills/`** (global).

```
.cursor/
├── config.json       # Paths and parameters for agents/scripts
├── hooks.json        # Event hooks
├── agents/           # Subagents (one .md per role)
├── commands/         # Commands (one .md per command; invoke as /filename)
├── rules/            # Rules (.mdc)
├── skills/           # Skills (folder/SKILL.md)
├── templates/        # Templates for commands
├── scripts/          # Scripts invoked from commands
└── hooks/            # Scripts for hooks.json (optional)
```

---

## 2. Subagents (agents)

**Location:** `.cursor/agents/<role>.md`  
**How to create:** In chat with the agent, run **/create-subagent** and describe the role — Cursor creates the file in the correct format (name, description, skill bindings).

**Invocation:** Subagents are not invoked directly by the user — they are launched by another agent via **mcp_task** with `subagent_type` = filename without `.md` (e.g. `worker`, `reviewer`). In commands (`.cursor/commands/`), steps must specify a call to `mcp_task(subagent_type="<role>", prompt="...", description="...")`.

---

## 3. Commands (commands)

**Location:** `.cursor/commands/<command-name>.md`  
**Invocation in Cursor:** `/command-name` (filename without `.md`).

**Format:** Markdown. Typical structure:

1. Title and short description of the command.
2. Usage (e.g. `/command-name arguments`).
3. **Steps** — step-by-step scenario. For steps that need a subagent role, explicitly state a **mcp_task** call with `subagent_type`, `prompt`, `description`; do not perform that role “as the orchestrator”.
4. Result — what to return to the user.
5. Notes (limits, parallel calls, etc.).

Commands may reference `.cursor/templates/` and parameters from `.cursor/config.json`.

**How to add a command:**

1. Create `.cursor/commands/<command-name>.md`.
2. Describe purpose, usage, and steps.
3. In steps that need a subagent — write `mcp_task(subagent_type="...", prompt="...", description="...")`.

---

## 4. Rules (rules)

**Location:** `.cursor/rules/` (**.mdc** files).  
**How to create:** In chat with the agent, run **/create-rule** and describe the rule (code style, commits, when to apply, etc.) — Cursor creates the rule file in the correct format.

---

## 5. Built-in Cursor commands: creating agents, rules, and skills

To create a subagent, rule, or skill, use the corresponding command in chat with the agent (type `/` and select the command). Cursor creates the files in the right format.

| Command | Creates | Saves to |
|---------|---------|----------|
| **/create-subagent** | Subagent (role for mcp_task) | `.cursor/agents/<role>.md` |
| **/create-rule** | Rule for agent context | `.cursor/rules/` (.mdc file) |
| **/create-skill** | Skill (instructions/workflow) | `.cursor/skills/<name>/SKILL.md` or `.agents/skills/` |
| **/migrate-to-skills** | (Cursor 2.4+) Migrate existing slash commands and dynamic rules into skills | `.cursor/skills/` |

**How to invoke:** In chat with the agent, type `/` and choose the command by name. List of skills and rules: **Cursor Settings → Rules**.

---

## 6. Skills (skills)

**Location:** One skill = folder with **SKILL.md**. Cursor loads skills from:

- **`.agents/skills/<skill-name>/`** and **`.cursor/skills/<skill-name>/`** — project
- **`~/.cursor/skills/<skill-name>/`** — global

**How to create:** In chat with the agent, run **/create-skill** and describe what the skill should do — Cursor creates the folder and SKILL.md in the correct format.

**How to use:** In chat type **/skill-name** (explicit call) or **@skill-name** (attach as context). By default the agent may apply the skill by relevance; to restrict to “only via /name” configure this in the skill’s SKILL.md (frontmatter option).

**Inside the skill folder** you can optionally add: **scripts/** (executable scripts), **references/** (extra docs), **assets/** (templates, configs). Standard: [agentskills.io](https://agentskills.io/).

---

## 7. Invoking subagents (mcp_task)

Subagents are invoked only via the **mcp_task** tool.

**Parameters:**

| Parameter      | Description |
|----------------|-------------|
| subagent_type  | Role name = filename in `.cursor/agents/` without `.md` |
| prompt         | Task text for the subagent (context, what’s already done) |
| description    | Short description of the call (3–5 words), for logs |
| resume         | (optional) ID of a previous agent to continue the session |

**Example:** `mcp_task(subagent_type="worker", prompt="Implement function X in file Y", description="Worker implementation")`.

Prefer passing context in the prompt (task, affected files, issue link) so the subagent doesn’t duplicate work.

---

## 8. config.json

**Location:** `.cursor/config.json`

Defines paths and parameters used by commands and scripts. Structure is project-specific; typical sections:

- **testing** — how to run tests (`command`, `path`) if a test-runner subagent reads config.
- **documentation** — paths for plans, reports, etc. (`paths`, `enabled`).
- **metrics** — where to save session reports (`sessionsPath`, `enabled`).

Commands and agents use these fields by convention (e.g. “save session report to config.metrics.sessionsPath”).

---

## 9. Hooks (hooks.json)

**Location:** `.cursor/hooks.json`

**Structure:**

```json
{
  "version": 1,
  "hooks": {
    "afterFileEdit": [],
    "subagentStop": [],
    "beforeShellExecution": [],
    "afterShellExecution": [],
    "sessionStart": [],
    "sessionEnd": []
  }
}
```

Each array holds actions for the event:

- **afterFileEdit** — after the agent edits a file.
- **subagentStop** — after a subagent finishes.
- **beforeShellExecution** / **afterShellExecution** — before/after running a shell command.
- **sessionStart** / **sessionEnd** — session start/end.

**Array element:** command string or object `{ "command": ".cursor/hooks/script.sh" }`. For LLM checks: `{ "type": "prompt", "prompt": "...", "timeout": 10 }`.

---

## 10. Templates and scripts

- **.cursor/templates/** — any templates (instructions for gh, report format, etc.). Commands and agents reference them by path.
- **.cursor/scripts/** — scripts invoked by commands (e.g. reports, metrics). Paths and parameters can come from `config.json`.

---

## 11. Creation checklist

1. **Subagents** — in chat with the agent: **/create-subagent**, describe the role. Files go to `.cursor/agents/`.
2. **Rules** — **/create-rule**, describe the rule. Files go to `.cursor/rules/`.
3. **Skills** — **/create-skill**, describe workflow/instructions. Files go to `.cursor/skills/` or `.agents/skills/`.
4. **Commands** — create manually `.cursor/commands/<name>.md` for scenarios; in steps call subagents via **mcp_task**.
5. Configure `.cursor/config.json` for project paths and tests.
6. Optionally — `.cursor/hooks.json` and scripts in `.cursor/hooks/`.
7. If you already have slash commands or dynamic rules — **/migrate-to-skills** (Cursor 2.4+) migrates them to skills.

Summary: agent, rule, and skill are created with Cursor commands (**/create-subagent**, **/create-rule**, **/create-skill**). Orchestration commands live in `.cursor/commands/`; subagents are invoked from them via mcp_task.
