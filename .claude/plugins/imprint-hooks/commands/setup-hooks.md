---
name: setup-hooks
description: Create hookify rules for imprint build automation
---

# Setup Imprint Hookify Rules

Create the following hookify rules in `.claude/` directory. These rules enforce build conventions for the imprint project.

## Rules to Create

### 1. `.claude/hookify.block-imprint-build-location.local.md`

```markdown
---
name: block-imprint-build-location
enabled: true
event: bash
action: block
conditions:
  - field: command
    operator: regex_match
    pattern: go build.*/cmd/imprint
  - field: command
    operator: not_contains
    pattern: -o bin/imprint
---

**Incorrect build output location!**

The imprint binary must be built to `bin/imprint`.

Use: `go build -o bin/imprint ./cmd/imprint`
```

### 2. `.claude/hookify.block-example-build-names.local.md`

```markdown
---
name: block-example-build-names
enabled: true
event: bash
action: block
conditions:
  - field: command
    operator: regex_match
    pattern: go build.*\./examples/
  - field: command
    operator: not_contains
    pattern: -o examples/
---

**Example binary must have proper name!**

Build examples with matching binary names:
- `go build -o examples/screenshot-demo/screenshot-demo ./examples/screenshot-demo`
- `go build -o examples/text-demo/text-demo ./examples/text-demo`
- `go build -o examples/what-changed/what-changed ./examples/what-changed`
```

### 3. `.claude/hookify.block-go-dependency-version.local.md`

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

### 4. `.claude/hookify.block-stop-without-build.local.md`

```markdown
---
name: block-stop-without-build
enabled: true
event: stop
action: block
conditions:
  - field: transcript
    operator: regex_match
    pattern: Edit.*\.go|Write.*\.go
  - field: transcript
    operator: not_contains
    pattern: go build
---

**Go files were modified but not built!**

Run the build before completing:
`go build -o bin/imprint ./cmd/imprint`

For examples, also build:
- `go build -o examples/screenshot-demo/screenshot-demo ./examples/screenshot-demo`
- `go build -o examples/text-demo/text-demo ./examples/text-demo`
- `go build -o examples/what-changed/what-changed ./examples/what-changed`
```

### 5. `.claude/hookify.warn-parallel-agents.local.md`

```markdown
---
name: warn-parallel-agents
enabled: true
event: prompt
action: warn
conditions:
  - field: user_prompt
    operator: regex_match
    pattern: implement|build|create|add.*feature|develop
---

**Consider using parallel agents!**

When implementing plans with independent tasks:
- Split work into parallel tracks
- Launch agents simultaneously with multiple Task tool calls
- Only serialize tasks that have dependencies
```

## Instructions

1. Create each file listed above with the exact content shown
2. After creating all files, confirm they exist by listing `.claude/hookify.*.local.md`
3. Tell the user the hooks are ready and will activate immediately (no restart needed)
