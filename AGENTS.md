# AGENTS.md

This file is created and edited by human. AI Agent should ask the user if want to modify this.

## Code navigation (CodeGraph + Serena)

- Use CodeGraph first for structural questions: callers/callees, blast radius, dependency/import maps
- Use Serena for symbol-level reads/edits: get_symbol_body, replace_symbol_body, rename_symbol
- Serena: activate at session start, and again before any Serena call if unsure it's still active (e.g. after a context reset)
- Serena: onboarding runs once per project — skip if project memories already exist
- Don't grep or read whole files for things either tool already answers directly

## Token tooling

- Caveman governs prose/chat replies only — never let it touch code comments or docs
- Ponytail governs code decisions (YAGNI ladder) — takes precedence in .go files

## Context7 usage flow

Trigger: any time the agent needs current docs for a library/framework/API (version-specific behavior, method signatures, config options) or prompt contains `use context7`

1. Normalize the query to a cache key: `<library>-<version>-<slugified-topic>` (e.g. `fiber-v3-middleware-order`).
2. Check `.context7/<cache-key>.md`. If it exists and is under 30 days old, use it — end.
3. Otherwise, query the `context7` MCP tool. Return the result, then write it to `.context7/<cache-key>.md` with a `fetched: <date>` header.

Note: `.context7/` is gitignored — local cache only, not shared across the team.

## Code lint quality

- **always** avoid error is not checked / unchecked errcheck
- **always** use tagged switch instead of `if` (QF1003) if possible

## JSON request/response

- **style** : camelCase

## MongoDB

- **database name** : snake_case
- **collection name** : snake_case
- **document key** : camelCase

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:ca08a54f -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd dolt push
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->
