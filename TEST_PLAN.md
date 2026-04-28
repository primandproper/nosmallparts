# No Small Parts — Pre‑Launch Test Plan

> Soft-launch (MVP) acceptance checklist. Desktop web only for v1. Mobile is **out of scope** for this pass — the framework supports it, but we are not certifying iOS/Android.

## Legend
- `[ ]` to do
- `[~]` partial / in progress
- `[x]` done
- `[!]` blocker found
- `[blocked: needs impl]` test depends on functionality not yet built — see the **Implementation Gaps** appendix

---

## Glossary & design notes
- **Identity** = either an authenticated `userId` (Better Auth) or a `guestId` (ULID stored in the `nsp_guest_id` cookie). Some flows require an authenticated user (creating a stage, hosting, rating, owning scripts).
- **Host** = the privileged identity on a stage. Set on stage creation; can transfer; on disconnect, a 15s grace then oldest-joined successor.
- **Director mode** = host-only flag toggled in lobby. When on: only the director advances lines; host cannot also be in the cast.
- **Speaker** = the participant currently assigned to `play.lines[currentLineIndex].castMemberId`.
- **Audience** = participants flagged `isAudience: true`. They cannot hold roles.
- **Admin** = a Better Auth user whose email is in the `ADMIN_EMAILS` env var (comma-separated, case-insensitive).
- **Open design question:** "host" vs "director" may eventually collapse. For v1 they remain separate.

---

## 0. Setup matrix
- [ ] Chrome (latest stable)
- [ ] Safari (latest stable)
- [ ] Firefox (latest stable)
- [ ] Two distinct browsers on one machine (host + guest in separate sessions)
- [ ] Two desktops on the same LAN (real WS, real LiveKit)
- [ ] Clean DB: `pnpm db:migrate && pnpm db:seed`
- [ ] `.env` with `LIVEKIT_*` set (full path)
- [ ] `.env` without `LIVEKIT_*` (verify graceful fallback message; rest of stage works)
- [ ] `BETTER_AUTH_SECRET` rotated mid-test does not crash existing clients (just signs them out)
- [ ] `ADMIN_EMAILS` env var honored (comma-separated, case-insensitive match against Better Auth user email)
- [ ] Server runs behind a reverse proxy with `X-Forwarded-For` set; verify client IP captured for ban/rate-limit purposes

---

## 1. Identity & Auth

### 1.1 Guest identity (`/api/identity`, `nsp_guest_id` cookie)
- [ ] Fresh visitor with no cookie can land on Home (server mints guest cookie)
- [ ] Guest cookie persists across reload and across tabs
- [ ] Clearing cookies starts a brand-new guest identity
- [ ] Tampering with cookie to a non-ULID is treated as no identity
- [ ] Native fallbacks (`?guest=`, `?token=`) — **not tested for v1** (mobile out of scope)

### 1.2 Email/password (Better Auth)
- [ ] Register with new email → signed in, redirected, session present at `/api/me`
- [ ] Register with existing email → friendly error, no partial account created
- [ ] Register with weak/invalid password → validation error surfaced in UI
- [ ] Login with correct creds → session cookie set
- [ ] Login with wrong creds → error, no session
- [ ] Sign out clears session; protected routes (`POST /api/play-scripts`, `POST /api/stages`, etc.) return 401
- [ ] Session survives reload; expired session degrades to guest cleanly
- [ ] Two browsers signed into the same account share scripts and dedupe to a single participant in any given stage

### 1.3 Guest → user upgrade
- [ ] After registering/logging in mid-session, the existing `nsp_guest_id` cookie is **kept** (not migrated, not cleared)
- [ ] No participant rows are migrated from the guest identity to the new user identity (intentional; document as known papercut)
- [ ] **[blocked: needs impl]** If a guest is somehow inside a stage and logs in, host status does **not** carry over (and per §1.4 a guest cannot be host anyway, so this should be unreachable)

### 1.4 Login required to host  **[blocked: needs impl]**
- [ ] `POST /api/stages` requires an authenticated user; guest gets 401
- [ ] WS join when stage has no host **does not** assign a guest as host
- [ ] If the only participants on a stage are guests, the lobby shows a clear "Waiting for a signed-in user to host" banner and start-play is disabled
- [ ] Host succession only considers identities with a `userId` (guests are skipped during election)
- [ ] If the host leaves and only guests remain, the stage clears the host and waits for a signed-in user to take over

### 1.5 Auth gates (HTTP)
- [ ] `POST /api/play-scripts` requires user (401 for guest)
- [ ] `PUT /api/play-scripts/:id` only allowed for the script's `createdByUser`; non-owner → 404
- [ ] `DELETE /api/play-scripts/:id` **admin-only** (see §2.4) — non-admin gets 403, owner non-admin also gets 403
- [ ] `POST /api/play-scripts/:id/rate` requires user; guest → 401
- [ ] `POST /api/livekit/token` issued only to participants currently joined to the stage, with audience receiving subscribe-only tokens

---

## 2. Scripts (CRUD, listing, ratings, tags, enabled flag)

### 2.1 Browse
- [ ] Anonymous visitor sees only public **enabled** scripts
- [ ] Logged-in user sees public enabled + their own enabled scripts (their disabled scripts are hidden from them too — only admins see disabled)
- [ ] Admin sees all scripts including disabled, with a visible "disabled" badge
- [ ] `?search=` filters case-insensitively across title/description/author
- [ ] User's own ratings reflected in `userRating` on each play

### 2.2 Create / edit
- [ ] Create script with cast + lines + tags (tag IDs and tag names) → persisted, retrievable
- [ ] New scripts default to `enabled = true`
- [ ] Edit own script — title, cast, lines, public toggle persist
- [ ] Edit script you don't own → 404
- [ ] Reordering / removing cast members and lines preserves index ordering
- [ ] `scrollMode: 'auto'` + `autoScrollWpm` actually drive auto-scroll on stage during performance

### 2.3 Ratings  **[blocked: needs impl for self-rating block]**
- [ ] Rate 1–5 → average and count update; user's own rating shown
- [ ] Re-rate replaces previous rating, count does not double
- [ ] Rate values outside 1–5 rejected with 400
- [ ] Authors **cannot** rate their own scripts (server returns 403)
- [ ] Guests cannot rate (server returns 401); UI on Discussion screen prompts unsigned-in users to "Sign in to rate"

### 2.4 Admin-only delete + `enabled` flag  **[blocked: needs impl]**
- [ ] DB has `enabled BOOLEAN NOT NULL DEFAULT TRUE` on `play_scripts`
- [ ] Admin endpoint to flip `enabled` on any script (or use existing PUT, gated)
- [ ] Non-admin DELETE → 403
- [ ] Admin DELETE → 200, script removed
- [ ] Disabling a script: new stage creation against that `playScriptId` is rejected (400 / "script unavailable")
- [ ] Disabling a script: live stages currently performing it **continue uninterrupted** until they end naturally
- [ ] Disabling a script: it disappears from `GET /api/play-scripts` for non-admins
- [ ] Re-enabling a disabled script restores it to listings
- [ ] Admin status is derived from `ADMIN_EMAILS` env var at request time (no DB column), case-insensitive

### 2.5 Tags
- [ ] `GET /api/tags` lists all tags
- [ ] Creating a script with `tagNames` upserts new tags
- [ ] Tag names normalized (case, trim) so duplicates merge

---

## 3. Stage lifecycle (lobby → countdown → performing → discussion)

### 3.1 Create / join
- [ ] Logged-in user creates stage with no script preselected → lobby loads, code visible, they are host
- [ ] Logged-in user creates stage with `?s=<scriptId>` → that play preselected
- [ ] Logged-in user cannot create a stage against a disabled script (§2.4)
- [ ] Joining via 6-char code (case-insensitive) lands in the right stage
- [ ] Joining a non-existent code shows "Stage not found" inline error
- [ ] Joining without entering a name is blocked
- [ ] Stage code generation never produces ambiguous chars (0/O/1/I/L)
- [ ] `MAX_PARTICIPANTS_PER_STAGE = 12` enforced; 13th gets "Stage is full" and is closed
- [ ] Display name validated on join (see §12.4)

### 3.2 Same identity, multiple connections
- [ ] Same user in two tabs: second tab displaces first; first receives `sessionReplaced` and closes; participant list still shows them once
- [ ] Same guest cookie in two tabs: identical behavior to logged-in case
- [ ] Reconnect after refresh reuses the existing participant row; cast roles preserved

### 3.3 Lobby actions
- [ ] Click cast member → claim role; click again → release
- [ ] Claiming a role another participant holds steals it (exclusive across participants)
- [ ] **Same participant can hold multiple roles** (intentional doubling)
- [ ] Toggling "audience" clears all of your roles
- [ ] Selecting a different play clears all roles for everyone
- [ ] `Ready` toggle visible to others in real time

### 3.4 Start-play preconditions  **[blocked: needs impl]**
- [ ] In non-director mode, `startPlay` is rejected unless **every cast role is filled**
- [ ] In non-director mode, `startPlay` is rejected unless **every non-audience participant has `isReady: true`**
- [ ] In director mode, only the director's `startPlay` is accepted (others get an error toast)
- [ ] In director mode, the same "all roles filled" + "all non-audience ready" preconditions still apply
- [ ] UI on the lobby disables the Start button when preconditions fail and shows what's missing ("3 unfilled roles, 2 not ready")

### 3.5 Countdown → Performing
- [ ] `countdownStart` shows a 3-second countdown to all clients within ~200 ms of each other
- [ ] After countdown, `playStarted` arrives; clients render Stage with line 0 highlighted
- [ ] Audience members see the same play view but no advance affordance

### 3.6 Line advancement  **[partially blocked: needs impl for speaker-only]**
- [ ] Default mode: only the **speaker of the current line** can advance (server rejects others with "Not your line")
- [ ] In `directorMode`, only host advances; non-host attempts get the gated error toast
- [ ] Audience cannot advance in any mode
- [ ] Reaching the last line transitions to `discussion` and emits `playEnded`
- [ ] Heartbeats every 5s reconcile any missed `lineAdvanced`

### 3.7 Pause / resume / back / skip / end
- [ ] Host pause → broadcast system message; `nextLine` rejected with "paused" error
- [ ] Host resume → flow continues from same line index
- [ ] Host back: clamps at 0; non-host → "Host only"; only allowed during `performing`
- [ ] Host skip: same gating as back; skipping past last line ends the play
- [ ] Host end: confirmation modal mid-scene; no confirm in `discussion`; resets to lobby with everyone un-ready

### 3.8 Discussion
- [ ] Cast list shown with assigned actor
- [ ] Logged-in users can rate the script from this screen
- [ ] Guests see a "Sign in to rate" prompt instead of stars
- [ ] Authors see their own rating disabled with a tooltip ("authors can't rate their own work")
- [ ] "Perform again" returns to lobby preserving selected play; everyone un-ready

### 3.9 Late joiners  **[blocked: needs impl]**
- [ ] Joining during `countdown`, `performing`, or `discussion` automatically lands in **audience** (`isAudience: true`)
- [ ] Late joiner cannot self-toggle out of audience until the stage returns to `lobby`
- [ ] Late joiner sees current line index in sync immediately
- [ ] System message announces the late join

---

## 4. Host & role management

### 4.1 Host election & succession
- [ ] First **signed-in** identity to join a stage with no host becomes host
- [ ] Guests joining a host-less stage do **not** become host (§1.4)
- [ ] If host disconnects, a 15s grace window is held; system does not announce succession yet
- [ ] If host returns within 15s, "host reconnected" message; no transfer
- [ ] If host does not return, oldest-joined remaining **signed-in** identity is promoted; system message names them
- [ ] If host is the only signed-in user and leaves, host is cleared; stage waits for a signed-in user
- [ ] Two tabs: closing one tab (host's) does NOT trigger succession because the other tab is still connected

### 4.2 `hostTransferHost`
- [ ] Host can transfer to any connected **signed-in** participant; old host becomes regular
- [ ] Transfer to a guest is rejected ("Target must be signed in")
- [ ] After transfer, host-gated actions work for new host and fail for old host

### 4.3 Reassign / unassign roles in lobby
- [ ] In lobby: reassigning role moves it instantly, system message posted
- [ ] Reassigning a role to someone who already holds another keeps both (doubling)
- [ ] Host unassign clears the role

### 4.4 Reassign mid-performance (manual host action)
- [ ] In performing: `hostReassignRole` pauses the play, sends `roleOffer` to target, posts system message
- [ ] Target accepts → role moves, system message, host must manually resume
- [ ] Target declines → role unchanged, play remains paused
- [ ] Target disconnects before responding → offer cleared, system message, play remains paused

### 4.5 Role vacancy mid-performance (cast member leaves)  **[blocked: needs impl]**
- [ ] Non-director mode: when the speaker (or any role-holder) disconnects mid-play:
  - [ ] Stage immediately enters `pausedByHost = true`
  - [ ] System message names the vacated role
  - [ ] A 45-second visible countdown starts (broadcast `roleVacated` with `claimDeadline` timestamp)
  - [ ] During the countdown, any audience member can claim the vacant role with a single click; on claim, the play remains paused but the role is filled and host can resume
  - [ ] Host can fast-track via existing `hostReassignRole` offer flow at any time during the countdown (bypasses timer)
  - [ ] If the timer expires with no claim and no audience available, the play **auto-ends** to discussion with a system message ("Play ended: role X went vacant")
  - [ ] If the original holder reconnects within the countdown, they reclaim the role and the countdown is cancelled
- [ ] Director mode: when a role vacates:
  - [ ] Stage pauses
  - [ ] Director uses the existing `hostReassignRole` offer flow, targeting any audience member with cam+mic capability
  - [ ] Eligibility check: audience member must have published mic AND camera tracks (or accept the offer with a "this will enable your cam+mic" confirmation)
  - [ ] Target must explicitly confirm role assignment (existing `respondRoleOffer` accept path)
  - [ ] No 45-second countdown in director mode

### 4.6 Director mode
- [ ] Toggle only allowed in lobby
- [ ] Enabling director mode clears host's own cast assignments
- [ ] Host cannot claim a role while director mode is on (error surfaced)
- [ ] During performing, only director's `nextLine` advances; others get "Director controls line advance"
- [ ] Disabling director mode (only in lobby) allows host to claim roles again

### 4.7 Kick / ban
- [ ] Host kicks participant → target receives error, socket closed, system message; can rejoin freely
- [ ] Host bans participant → target same as kick; subsequent joins from same identity blocked with "removed" message
- [ ] **[blocked: needs impl]** IP ban: banned identity's IP (from `X-Forwarded-For`, last public hop) is recorded in `stage_ip_bans`; subsequent connections from that IP to this stage are rejected
- [ ] **[blocked: needs impl]** Banned guest who clears their cookie: still blocked because IP ban catches them
- [ ] **[blocked: needs impl]** Banned user logged out and joined as guest from a different IP: not blocked (acceptable per spec)
- [ ] Host cannot kick/ban self
- [ ] After kick, target's LiveKit participant is removed server-side
- [ ] IP bans are scoped per-stage, not global

### 4.8 Mute  **[blocked: needs impl for visual disable]**
- [ ] Host mutes target → `forceMuted` flag set, LiveKit unpublishes target's audio
- [ ] Target's local mic-toggle button is **visually disabled** with a tooltip ("muted by host")
- [ ] Host unmutes → button re-enabled
- [ ] Target cannot re-publish audio while force-muted (server rejects via LiveKit admin)

---

## 5. Inviting / sharing a stage

### 5.1 Invite link button
- [ ] "Copy invite link" button in `Lobby.tsx` copies a working URL to clipboard
- [ ] Pasting the URL in a new tab/browser lands the recipient on `/stage/:code`
- [ ] Toast/feedback "Invite link copied" appears on success
- [ ] On clipboard API failure (e.g., insecure context), a fallback shows the URL for manual copy

### 5.2 Logged-in user clicks invite link
- [ ] Lands directly on `/stage/:code`, prompted only for display name (or auto-uses Better Auth name)
- [ ] No registration prompt
- [ ] Joins as audience if stage is past lobby (§3.9), else can claim cast roles

### 5.3 Anonymous (no account) clicks invite link
- [ ] Lands on `/stage/:code`, prompted for a display name only — **not** required to register
- [ ] Server mints a guest cookie on first request
- [ ] Joins successfully as a guest (cannot host, can claim cast roles in lobby)
- [ ] A small "Sign in" affordance is visible but optional

### 5.4 Banned identity follows a fresh invite link
- [ ] Banned by identity → blocked with "removed" message
- [ ] Banned by IP → blocked even if they sign in or use a fresh guest cookie (§4.7)
- [ ] From a different IP, banned guest who registers a new account: **not blocked** (acceptable per spec)

### 5.5 Stage code privacy
- [ ] Stage codes don't appear in any analytics/telemetry events
- [ ] Stage codes don't appear in server logs at INFO level
- [ ] `/api/stages/:code/exists` is rate-limited (§12.5)

---

## 6. Real-time chat  **[partially blocked: needs impl for markdown + truncation]**

### 6.1 Visibility
- [ ] Public chat reaches everyone including audience
- [ ] Cast-only chat reaches only non-audience participants
- [ ] Toggling to/from audience changes whether you receive future cast-only messages
- [ ] System messages render distinctly (no sender, italic style)

### 6.2 Formatting
- [ ] Bold (`**x**`), italic (`*x*` and `_x_`), inline code (`` `x` ``), strikethrough (`~~x~~`) all render correctly
- [ ] Links (`[text](url)`) render as **plain text** (the markdown is shown literally or stripped — no clickable link)
- [ ] Bare URLs (`https://example.com`) are not autolinked
- [ ] Images (`![alt](url)`) render as plain text / stripped
- [ ] Headings (`# x`), blockquotes (`> x`), tables, raw HTML — all stripped/escaped
- [ ] Emojis render correctly (native unicode, not custom)

### 6.3 Length & truncation
- [ ] A `MAX_CHAT_LENGTH` const exists (initial value: 256)
- [ ] Messages over `MAX_CHAT_LENGTH` are auto-truncated server-side before broadcast
- [ ] Truncated messages get a trailing ellipsis or indicator
- [ ] Client input does not hard-cap, only truncates on send (per spec: no length cap on input)

### 6.4 Persistence
- [ ] Chat history is **not** persisted; rejoining users do not see past messages
- [ ] System messages are also not persisted

---

## 7. Video / mic (LiveKit)

### 7.1 With LiveKit configured
- [ ] **Mic and camera both off by default** on join — no auto-publish
- [ ] User can opt in to mic; works
- [ ] User can opt in to camera; works
- [ ] Audience tokens are subscribe-only (cannot publish)
- [ ] Token TTL refresh works for long sessions (>1 hour)
- [ ] Network drop → reconnect restores published tracks without rejoining the stage room
- [ ] Up to 12 participants render without obvious distortion / dropped frames
- [ ] Force-mute (§4.8) actually unpublishes the target's audio track on the SFU

### 7.2 Without LiveKit configured
- [ ] App shows the friendly fallback message; rest of stage features still work
- [ ] No console errors or repeated token request attempts

---

## 8. Connection robustness
- [ ] Server restart while clients are connected → clients show "disconnected", auto-reconnect succeeds, state intact
- [ ] WebSocket survives Wi-Fi blip on desktop (drop interface for 10s, reconnect)
- [ ] Heartbeat reconciles missed `lineAdvanced`/`stageState` after suspension
- [ ] Disposing the stage tab fires `onClose` which schedules host succession correctly
- [ ] Sending malformed JSON → server returns "Invalid JSON" error, doesn't crash
- [ ] Sending unknown message `type` → "Unknown message type" error, connection stays open
- [ ] WS connection without prior identity is rejected at `join` time

---

## 9. Concurrency / edge cases
- [ ] Two participants click the same cast member at the same instant → exactly one ends up holding it; UI converges
- [ ] Host taps `Pause` then `End play` rapidly → state machine doesn't deadlock
- [ ] `selectPlay` mid-lobby clears roles for everyone (verify no orphan role rows)
- [ ] Host disconnects during a pending role offer → offer auto-cleared, play remains paused
- [ ] Two simultaneous `hostKick` requests for the same target → only one succeeds, other gets graceful error
- [ ] Late joiner during the 45s vacancy countdown sees the countdown remaining time, not a stale value
- [ ] Participant rejoins after `playEnded` and discussion already started → lands in discussion with cast list visible

---

## 10. Persistence, cleanup & lifetime  **[blocked: needs impl for cleanup]**

### 10.1 Persistence
- [ ] Restart server → existing stage codes still resolve (until cleanup runs); participants must rejoin
- [ ] `leftAt` is set on disconnect; `getStageState` filters by `isNull(leftAt)`
- [ ] Bans persist across server restarts (both identity and IP)
- [ ] Rating averages recompute correctly after delete/re-rate

### 10.2 Stage cleanup
- [ ] When the **last** participant leaves a stage, the stage record is deleted (or hard-tombstoned) within 5 seconds
- [ ] After cleanup, `GET /api/stages/:code/exists` returns `{exists: false}` — "as if it never existed"
- [ ] An idle stage with no participants for **10 minutes** is also deleted (background reaper)
- [ ] Cleanup also removes: `stage_participants`, `stage_participant_roles`, `stage_bans`, `stage_ip_bans`, in-memory connection map entries, pending offers, host-grace timers
- [ ] Reaper runs even if the server was restarted while the stage was idle (recovered from DB)
- [ ] Cleanup does not race with a new joiner: if someone joins during the 5s window, cleanup is cancelled

---

## 11. UX / polish (desktop)
- [ ] All error toasts dismissable; do not stack indefinitely
- [ ] Disabled-button visual state distinct (already wired in `HostControls.ActionButton`)
- [ ] Layout works at 1280×720 minimum and at ≥1920×1080
- [ ] Dark theme contrast meets WCAG AA on all key text
- [ ] Keyboard navigation: Tab cycles inputs/buttons sensibly
- [ ] Loading spinners present for any await > 200 ms (`api()` calls, WS connect)
- [ ] Empty states for: no scripts in browse, no participants in lobby (just you)
- [ ] "Copy invite link" button is prominent in lobby
- [ ] Force-mute disabled mic icon is unambiguous

---

## 12. Security / abuse / rate limiting

### 12.1 CSRF / CORS
- [ ] State-changing endpoints reject cross-origin requests without auth credentials
- [ ] CORS allowlist matches deploy domains (no `*` in production)
- [ ] WS upgrade respects same auth/identity as HTTP

### 12.2 Auth endpoints
- [ ] Better Auth's built-in rate limiting on sign-in / register is enabled
- [ ] Account enumeration on register is mitigated (consistent error timing & wording)

### 12.3 Logging hygiene
- [ ] No PII in server logs (emails, tokens, full session IDs)
- [ ] No stage codes in INFO logs

### 12.4 Display name policy  **[blocked: needs impl]**
- [ ] Length cap: **32 characters** (server-side; UI prevents over-typing)
- [ ] Profanity filter via `bad-words` (or equivalent) — rejected names show "Please choose a different name"
- [ ] Reserved tokens (case-insensitive): "Host", "Admin", "Director", "System", "Server" — rejected
- [ ] Per-stage uniqueness: case-insensitive; on collision the server **auto-suffixes** `(2)`, `(3)`, etc.
- [ ] Validation runs on initial join AND on rename (if rename is exposed)

### 12.5 Stage code lookup rate limit  **[blocked: needs impl]**
- [ ] `GET /api/stages/:code/exists` limited to **30 requests / minute / IP**
- [ ] Overage returns 429 with `Retry-After`
- [ ] Limit is per-IP, applied at the proxy or in middleware
- [ ] Limit is documented in `README.md` for ops

### 12.6 IP capture trust
- [ ] Server uses `X-Forwarded-For` last public hop for IP ban + rate limiting
- [ ] If proxy is misconfigured, document loud failure mode (don't silently ban deployer's IP)

---

## 13. Deployment smoke
- [ ] `pnpm dev` clean run: server :3001 + Expo :8081 (web only used)
- [ ] `pnpm db:migrate` from a fresh SQLite file applies all migrations cleanly
- [ ] `pnpm db:seed` produces the documented sample plays
- [ ] Production build of web app loads against deployed server
- [ ] `EXPO_PUBLIC_API_URL` override honored
- [ ] `/api/health` returns 200 in prod
- [ ] `ADMIN_EMAILS` set in prod env
- [ ] LiveKit env vars set in prod (or intentionally absent)

---

## Appendix A: Implementation Gaps

The following functionality is referenced by this plan but not yet built. Each is a prerequisite for the corresponding `[blocked: needs impl]` checks. Suggest filing as separate issues.

1. **Auth-required stage creation** (§1.4)
   - Add `requireAuth` to `POST /api/stages`
   - In `stage-handler.ts`, skip guest identities during host election
   - In `pickSuccessorIdentity`, exclude guest-only identities
   - Add lobby banner copy for "waiting for host"

2. **Self-rating block** (§2.3)
   - In `ratePlayScript`, return 403 if `play.createdByUser === userId`
   - Discussion screen disables the stars for authors with tooltip

3. **Admin role + `enabled` flag on plays** (§2.4)
   - Migration: `ALTER TABLE play_scripts ADD COLUMN enabled BOOLEAN NOT NULL DEFAULT TRUE`
   - Admin check helper reading `process.env.ADMIN_EMAILS` (comma-split, case-insensitive)
   - Gate `DELETE /api/play-scripts/:id` on admin
   - `listPlayScripts`: filter `enabled = true` for non-admins
   - `getPlayScript` for stage creation: reject when `enabled = false` (don't break live stages)
   - Optional admin endpoint `PATCH /api/play-scripts/:id/enabled`

4. **Start-play preconditions** (§3.4)
   - In `startPlay` WS handler: validate every cast role has a holder; validate every non-audience participant `isReady`
   - In director mode: only the host's `startPlay` is honored
   - Lobby UI shows live preconditions feedback

5. **Speaker-only line advancement** (§3.6)
   - In `nextLine` handler, when not in director mode, verify `participantId` holds `play.lines[stage.currentLineIndex].castMemberId`; otherwise reject with "Not your line"

6. **Late-joiner audience lock** (§3.9)
   - In `join` handler, if `stage.phase !== 'lobby'`, set `isAudience: true` on the new participant row and reject `setAudience(false)` until phase returns to `lobby`

7. **Mid-play role vacancy flow** (§4.5)
   - On `onClose` for a role-holder during `performing`: pause stage, broadcast `roleVacated { castMemberId, claimDeadline }`, start a per-stage timer
   - New WS message: `audienceClaimRole { castMemberId }` — gated to current audience and to the unfilled role
   - Director-mode branch: skip countdown, require host offer flow with mic+cam check
   - On timer expiry with no claim: transition to `discussion` with `playEnded` and a system message
   - On original holder's reconnect: cancel timer, restore role
   - Cancel timer on host fast-track reassignment

8. **IP ban capture** (§4.7)
   - Capture `X-Forwarded-For` (last public hop) at WS upgrade; store on `stageIpBans` row
   - Schema: `CREATE TABLE stage_ip_bans (id, stage_id, ip, banned_at)`
   - On WS join, look up by `(stage_id, ip)` in addition to identity ban check

9. **Force-mute UI disable** (§4.8)
   - Wire `forceMuted` on participant into mic-toggle disabled state in `VideoRoom.tsx`
   - Tooltip "Muted by host"

10. **Markdown chat + length truncation** (§6.2, §6.3)
    - Add a markdown renderer (e.g., `marked` with a custom renderer that drops links/images/headings/blockquotes/tables/HTML) or a small allowlist transform
    - Const `MAX_CHAT_LENGTH = 256`; truncate server-side in `sendChat` handler

11. **Stage cleanup** (§10.2)
    - On `onClose`, after removing connection: if `connections.get(stageId)?.size === 0` and DB shows no other live participants, delete stage and dependents
    - Background interval (e.g., every 60s): delete stages whose latest `leftAt`/`createdAt` is older than 10 minutes with no live connections
    - Ensure host-grace timer / pending offers / etc. are also cleared

12. **Display-name policy** (§12.4)
    - Validation function: length ≤ 32, not in reserved set, profanity filter
    - On join: check per-stage uniqueness (case-insensitive); auto-suffix on collision
    - Apply to both initial join and rename paths

13. **Stage code lookup rate limit** (§12.5)
    - Middleware on `GET /api/stages/:code/exists`: 30/min/IP, 429 + `Retry-After` on overage

---

## Appendix B: Open design questions
- **Host vs Director consolidation.** v1 keeps them separate. Revisit for v2: collapse "host" into "director" so there is exactly one privileged user per stage and director mode becomes the only mode.
- **Guest → user upgrade migration.** v1 papercut: guest who logs in inside a stage doesn't carry their participant row. Acceptable for soft launch.
- **Banned-identity bypass.** v1 accepts that a banned guest who registers a new account from a different IP can rejoin. Per-stage bans are short-lived enough that this is fine.
