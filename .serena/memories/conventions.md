# Conventions

From `AGENTS.md`:

- **Always** check errors — no unchecked errcheck violations
- **Use tagged switch** over cascading `if` (QF1003)
- Ponytail governs code decisions (YAGNI ladder)
- Caveman governs chat replies only, not code/docs

## Go conventions (planned)

- Standard Go project layout: `cmd/`, `internal/`, `pkg/`
- Dependency injection via constructor functions
- MongoDB repository pattern
- Structured JSON error responses matching the PRD contract
