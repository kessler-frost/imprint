# Imprint - Claude Code Instructions

## Binary Location

The MCP server binary is at `bin/imprint`. After building, always copy to this location:
```bash
go build -o bin/imprint ./cmd/imprint
```

## Dependencies

Always use the latest stable versions of dependencies:
- **Go**: Use the latest stable Go version (currently 1.25.x as of Dec 2025)
- **go-rod**: Use latest stable version
- **mcp-go**: Use latest stable version

When adding new dependencies, check for the latest version first using `go list -m -versions <package>`.

## Code Style

- Minimize if/else conditions and try/except blocks - avoid multiple code paths
- Use `pathlib` equivalent patterns when dealing with file/directory paths
- Keep code simple and focused on the task at hand

## Examples

Build examples with their proper binary names:
```bash
go build -o examples/screenshot-demo/screenshot-demo ./examples/screenshot-demo
go build -o examples/text-demo/text-demo ./examples/text-demo
go build -o examples/what-changed/what-changed ./examples/what-changed
```

## Plan Implementation

When implementing plans, **always use multiple agents in parallel** where tasks are independent:
- Split work into parallel tracks (e.g., Agent 1: rename/refactor, Agent 2: create new code)
- Launch agents simultaneously in a single message with multiple Task tool calls
- Only serialize tasks that have dependencies on each other
