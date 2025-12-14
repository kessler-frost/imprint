---
name: setup-hooks
description: Create hookify rules for imprint build automation
---

# Setup Imprint Hookify Rules

Create the following hookify rules in `.claude/` directory. These rules enforce build conventions for the imprint project.

## Rules to Create

### 1. `.claude/hookify.block-go-dependency-version.local.md`

```markdown
---
name: block-go-dependency-version
enabled: true
event: bash
action: block
conditions:
  - field: command
    operator: regex_match
    pattern: go get\s+
---

**Check latest version before adding dependency!**

Before running `go get`, first check the latest version:
`go list -m -versions <package>`

Then add the specific version you want.
```

### 2. `.claude/hookify.warn-parallel-agents.local.md`

```markdown
---
name: warn-parallel-agents
enabled: true
event: prompt
action: warn
conditions:
  - field: user_prompt
    operator: regex_match
    pattern: implement|build|create|add.*feature|develop|plan
---

**Plan for parallel agent execution!**

When writing this plan:
- Identify independent tasks that can run in parallel
- Group tasks by dependencies - independent tasks should be in the same phase
- Explicitly note which tasks can be launched simultaneously via multiple Task tool calls
- Only serialize tasks that depend on outputs from previous tasks
```

## Instructions

1. Create each file listed above with the exact content shown
2. After creating all files, confirm they exist by listing `.claude/hookify.*.local.md`
3. Tell the user the hooks are ready and will activate immediately (no restart needed)
