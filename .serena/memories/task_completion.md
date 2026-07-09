# Task Completion Checklist

From `AGENTS.md`:

1. File beads issues for remaining work
2. Run quality gates (if code changed): `go test ./...`, `go fmt ./...`
3. Update bd issue status (`bd close <id>` for finished, `bd update --notes="..."` for in-progress)
4. Push: `git pull --rebase`, `bd dolt push`, `git push`
5. Verify: `git status` must show "up to date with origin"

**Critical**: work NOT complete until `git push` succeeds.
