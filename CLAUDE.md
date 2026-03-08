# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Satirical web app tracking power outages in Santa Eulàlia de Ronçana. Stack: React 18 + TypeScript + Tailwind (frontend) + Go serverless functions (backend) + Neon Postgres (DB), deployed on Vercel.

Communication
Always communicate with the user in Spanish, the language used in code, comments, or commit messages has to be always in English.

## Testing — MANDATORY RULE

**Tests are non-negotiable.** Every new feature or bug fix MUST include unit tests. No implementation is considered done without passing tests.

### Rules

1. **Write tests before marking a task complete.** If you add or change logic, you add or change tests.
2. **Run all tests after every implementation** — both frontend and backend. Do not stop at build passing.
3. **Tests must pass with zero failures.** A red test suite blocks the task.
4. If an existing test breaks due to your changes, fix the test (or the code) before finishing.

### Frontend tests

```bash
npm test                # run all tests once (CI mode)
npm run test:watch      # watch mode during development
```

Test files live next to their component: `src/components/__tests__/Foo.test.tsx`

Framework: **Vitest + React Testing Library + jsdom**

### Backend tests (Go)

Because vercel-go compiles each `.go` file in isolation, test each file independently:

```bash
cd api
go test weather.go weather_test.go -v
go test stats.go   stats_test.go   -v
go test report.go  report_test.go  -v
```

**Never run `go test ./...`** — the duplicate helper functions across handlers cause compile errors.

Test files: `api/*_test.go` — same package (`package handler`), table-driven tests.

### What to test

| Layer | Test | Method |
|---|---|---|
| Go pure functions | `computeStats`, `parseWeatherAlert`, `sha256hex`, `clientIP` | Table-driven unit tests |
| React components | Render output, conditional display, fetch mocking, user events | RTL + `vi.fn()` |
| API integration | Not tested locally — covered by vercel-go compilation + manual `vercel dev` |

### What NOT to test

- Functions that only wrap DB or HTTP calls (`openDB`, the Handler itself) — mock or skip
- Styling/CSS classes — test behaviour, not class names

---

## Commands

## Skills

Before writing any code, **read** the corresponding SKILL.md file and **print the skill name visibly** in the response using the format `[skill: <name>]`. This must appear before any code or implementation detail.

| Situation               | Skills to load                                                                     | Path                                                                                                                                     |
|-------------------------|------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------|
| Backend changes         | `golang-pro`                                                                       | `.claude/skills/golang-pro/SKILL.md`                                                                                                     |
| Database changes | `neon-postgres`                                                                    | `.claude/skills/neon-postgres/SKILL.md`                                                                                                  |
| Frontend / UI changes   | `vercel-react-best-practices` **+** `ui-ux-pro-max` **+** `tailwind-design-systems` | `.claude/skills/vercel-react-best-practices/SKILL.md` · `.claude/skills/ui-ux-pro-max/SKILL.md` · `.claude/skills/tailwind-design-system` |

`vercel-react-best-practices` and `ui-ux-pro-max` are complementary and must always be loaded together for any frontend or UI work.

**This is mandatory and non-negotiable.** The user explicitly requires seeing `[skill: <name>]` printed in the response to verify correct skill usage.


### Frontend
```bash
npm run dev        # Vite dev server (port 5173)
npm run build      # tsc + vite build → dist/
npm run preview    # preview the dist/ build
```

### Backend (Go)
```bash
cd api
go mod tidy                        # sync go.sum
go build report.go                 # validate report handler in isolation
go build stats.go                  # validate stats handler in isolation
go vet report.go && go vet stats.go  # vet each file independently
```

> **Important:** `go vet ./...` (or `go build ./...`) on the whole `/api` directory **intentionally fails** — vercel-go compiles each `.go` file as an isolated function. Always validate files individually.

> **Adding a new handler:** When creating a new `api/foo.go`, add it explicitly to the `functions` block in `vercel.json`. The glob `api/*.go` is intentionally NOT used to prevent Vercel from picking up `_test.go` files as serverless functions.

### Database
```bash
psql $DATABASE_URL -f schema.sql   # create tables (run once on Neon)
```

### Local full-stack
```bash
vercel dev   # runs both Vite frontend and Go functions with correct routing
```

## Architecture

### Multi-runtime Vercel project
- `vercel.json` routes `api/*.go` to the `@vercel/go@3.4.3` runtime; everything else falls through to the React SPA (`dist/`).
- There is **no shared code** between `api/report.go` and `api/stats.go` — each file is self-contained with its own helper functions (`openDB`, `setCORSHeaders`, `writeJSON`). This duplication is intentional because vercel-go compiles each file in isolation.

### Go handlers (`api/`)
Both files use `package handler` and export exactly `Handler(w http.ResponseWriter, r *http.Request)`.

| File | Endpoint | Responsibility |
|---|---|---|
| `report.go` | `POST /api/report` | Hash client IP (SHA-256), check `daily_votes` for duplicate, upsert `incidents`, insert `daily_votes` |
| `stats.go` | `GET /api/stats` | Query all incident dates for the current year, compute 5 statistics in Go (no SQL aggregation) |

Stats are computed in Go by iterating sorted dates: `computeStats(dates []time.Time, now time.Time)` calculates streaks and gaps without helper queries.

All timestamps use `Europe/Madrid` timezone. `DATABASE_URL` is read from `os.Getenv`.

### React frontend (`src/`)
- `App.tsx` — root layout: fixed SVG star field (seeded PRNG, stable across renders), hero counter, `BatSignal`, `Stats`. Passes `fetchStats` as `onSuccess` to `BatSignal` so stats refresh after a successful report.
- `BatSignal.tsx` — the core interactive component. The SVG `<g>` oscillates ±15° via CSS `animation: oscillate` with `transformOrigin: '150px 458px'` (projector centre). State machine: `idle → loading → success | already_voted | network_error | server_error → idle`. Cookie `afc_voted` stores today's date as a fast client-side duplicate check before hitting the API.
- `Stats.tsx` — renders 5 stat cards from `StatsResponse`; shows skeleton while loading.
- `ErrorPage.tsx` — 404 page, same SVG bat signal rendered dim/greyscale.

### CSS animations
Custom keyframes live in two places:
1. `tailwind.config.js` — opacity/transform animations (`oscillate`, `shake`, `pulse-beam`, `blink`, etc.) usable as Tailwind classes.
2. `src/index.css` — `@keyframes flash-letter` (animates SVG `fill`) and `.bat-letter-flash` class, because Tailwind cannot animate SVG fill properties.

### Data schema
Two tables in Neon Postgres (see `schema.sql`):
- `incidents (id, date UNIQUE, created_at)` — one row per blackout day.
- `daily_votes (ip_hash TEXT, date DATE, created_at) PK(ip_hash, date)` — deduplication table.

### Design tokens
- Accent / signal colour: `signal-500` = `#fcd34d` (amber) — defined in `tailwind.config.js` as custom `signal` palette.
- Fonts: **Anton** (headings, big counter, SVG letter "A") + **JetBrains Mono** (stats, labels).

### Task Tracking
Use bd (Beads) for task tracking:

bd new "title" — create an issue before starting any planned task
bd state <id> in_progress — mark in-progress when beginning (bd start does not exist)
bd close <id> — mark as done
bd ready — review what's pending at end of session
Update PROGRESS.md after each completed task so work can be resumed from scratch if the session is interrupted.