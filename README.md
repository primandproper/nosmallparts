# No Small Parts

Real-time collaborative theater — create scripts and perform them together with friends in sync. Full-stack TypeScript, runs on iOS, Android, and the web.

## Stack

| Layer      | Tech                                                                     |
| ---------- | ------------------------------------------------------------------------ |
| Monorepo   | pnpm workspaces                                                          |
| Client     | Expo (Expo Router) + React Native + NativeWind                           |
| Server     | Hono on Node (HTTP + WebSocket via `@hono/node-ws`)                      |
| Database   | SQLite via `better-sqlite3` + Drizzle ORM                                |
| Auth       | [Better Auth](https://better-auth.com) (email/password; SSO-ready)       |
| Real-time  | WebSocket (stage state + line advancement)                               |
| Video/mic  | LiveKit (optional; server mints tokens, client uses native + web SDKs)   |

## Packages

```
packages/
  shared/   @nsp/shared  Pure TS types, stage logic, script utilities, sample plays
  server/   @nsp/server  Hono API, Drizzle schema, Better Auth, WebSocket stage handler
  app/      @nsp/app     Expo Router app — home, auth, scripts, stage, video
```

## Getting started

```bash
pnpm install
pnpm db:migrate
pnpm db:seed      # seeds the built-in sample plays
pnpm dev          # starts server on :3001 and Expo on :8081 in parallel
```

Open <http://localhost:8081> for web, or press `i` / `a` in the Expo CLI to launch a native simulator.

### Environment variables (server)

Copy `packages/server/.env.example` to `packages/server/.env` and set at least:

- `BETTER_AUTH_SECRET` — any long random string
- `LIVEKIT_API_KEY`, `LIVEKIT_API_SECRET`, `LIVEKIT_URL` — optional; unset disables video with a friendly fallback
- `ADMIN_EMAILS` — comma-separated, case-insensitive list of admin emails. Admins can delete any script and toggle the `enabled` flag; no DB column, evaluated per request.

### Ops notes

- **Rate limiting:** `GET /api/stages/:code/exists` is limited to **30 requests / minute / IP** (in-process middleware). Overage returns `429` with a `Retry-After` header.
- **Client IP & reverse proxies:** the server derives the client IP from the **last** entry of `X-Forwarded-For` (falling back to the socket address) for per-stage IP bans and rate limiting. If the app runs behind a reverse proxy that does **not** set `X-Forwarded-For`, all clients appear to share the proxy's IP — a single host-issued ban or a burst of lookups would then block *everyone* (including the deployer). Make sure the proxy sets the header before going live.
- **Stage cleanup:** stages are hard-deleted ~5s after the last participant disconnects, and a background reaper removes stages idle for 10+ minutes (survives restarts; recovered from the DB).

### Environment variables (app)

- `EXPO_PUBLIC_API_URL` — override the server URL if not `http://<lan>:3001`

## Data model

SQLite tables, managed by Drizzle migrations in `packages/server/src/db/migrations/`:

- Better Auth: `user`, `session`, `account`, `verification`
- Content: `play_scripts`, `play_script_cast_members`, `play_script_lines`, `script_tags`, `play_script_tags`, `play_script_ratings`
- Real-time: `stages`, `stage_participants`, `stage_participant_roles`

## API

- `GET /api/health` — liveness
- `GET /api/me` — current session
- `POST /api/auth/*` — Better Auth (sign up, sign in, sign out, etc.)
- `GET|POST|PUT|DELETE /api/play-scripts` — script CRUD
- `POST /api/play-scripts/:id/rate` — 1–5 star ratings
- `GET /api/tags` — tag list
- `POST /api/stages`, `GET /api/stages/:code/exists` — stage HTTP
- `WS  /ws/stage` — real-time stage protocol (see `packages/shared/src/types.ts`)
- `POST /api/livekit/token` — mint scoped LiveKit tokens (audience = subscribe-only)

## Real-time protocol

Messages are typed in `@nsp/shared` (`ClientMessage` / `ServerMessage`). The lifecycle:

`lobby` (pick play, claim roles, toggle ready) → `countdown` (3s) → `performing` (director/speaker advances lines) → `discussion` (cast list, rate, perform again).

Heartbeats every 5s let clients reconcile missed updates.

## License

TBD
