# Imprint - Claude Code Instructions

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
