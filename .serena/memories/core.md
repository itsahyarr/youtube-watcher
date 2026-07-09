# YouTube Watcher - Core

Planned Go service that receives a YouTube URL via HTTP API, opens it in a browser (Rod), clicks play, and logs the result to MongoDB.

## Source map

- `PRD.md` — full product requirements, API contract, schema, flow
- No code exists yet; project is in planning phase

## Invariants

- One scrape attempt = one MongoDB log document (success or failure)
- Browser runs with `headless=false` in development
- Only YouTube URLs accepted (www.youtube.com, youtube.com, youtu.be)
- Single endpoint: `POST /api/v1/scrape/youtube/play`

## References

- Tech stack details: `mem:tech_stack`
- Dev commands once code exists: `mem:suggested_commands`
- Code conventions: `mem:conventions`
- Task completion checklist: `mem:task_completion`
