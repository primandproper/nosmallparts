# No Small Parts — Pre‑Launch Test Plan

> Soft-launch (MVP) acceptance checklist. Desktop web only for v1. Mobile is **out of scope** for this pass — the framework supports it, but we are not certifying iOS/Android.

## Legend
- `[ ]` to do
- `[~]` partial / in progress
- `[x]` done
- `[!]` blocker found
- `[blocked: needs impl]` test depends on functionality not yet built — see the **Implementation Gaps** appendix

> **Status note (2026-07-21):** every Appendix A implementation gap is now built. An automated
> integration suite lives in `packages/server/test/` (88 tests: `pnpm test` from the repo root) and
> covers the server-side behavior end-to-end (HTTP via Hono `app.request`, WS via a handler-level
> harness with real SQLite). In addition, a headless-Chromium E2E pass (Playwright, three concurrent
> browser contexts against `expo start --web` + the real server) exercised: register → create stage →
> lobby, reserved-name rejection, guest join by lowercase code, chat markdown rendering, role
> claim/ready/auto-start/countdown/performing sync, speaker-only advance UI, mid-play vacancy
> countdown + audience claim, discussion ratings (stars for signed-in non-authors, "Sign in to rate"
> for guests), and perform-again. Screenshots + driver scripts in the session scratchpad
> (`e2e/drive1.js`, `e2e/drive2.js`). `[x]` = verified by suite, smoke check, or that E2E pass;
> `[~]` = implemented but still needing a human pass (real LiveKit A/V, multi-machine, visual QA).

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
- [~] Chrome (latest stable) — headless Chromium engine verified via E2E; branded-Chrome eyeball pass still recommended
- [~] Safari (latest stable) — headless WebKit engine verified via E2E (join, lobby, chat); branded-Safari pass still recommended
- [~] Firefox (latest stable) — headless Firefox engine verified via E2E (join, lobby, chat); branded pass still recommended
- [x] Two distinct browsers on one machine (host + guest in separate sessions) — E2E ran Chromium host + Firefox and WebKit guests simultaneously on one stage; participant lists and chat converged across all three
- [ ] Two desktops on the same LAN (real WS, real LiveKit)
- [x] Clean DB: `pnpm db:migrate && pnpm db:seed` (verified against a fresh SQLite file; 4 sample plays seeded)
- [~] `.env` with `LIVEKIT_*` set — room signaling connects with the real creds (participant tiles render for all three clients), defaults verified; actual A/V publish couldn't complete in headless sandbox (WebRTC ICE) — needs one human pass
- [x] `.env` without `LIVEKIT_*` (server 503s; UI shows "Video disabled (LiveKit not configured)"; single token request, no retries — verified in E2E screenshots)
- [ ] `BETTER_AUTH_SECRET` rotated mid-test does not crash existing clients (just signs them out)
- [x] `ADMIN_EMAILS` env var honored (comma-separated, case-insensitive match against Better Auth user email) — `test/units.test.ts`, `test/http-scripts.test.ts`
- [x] Server runs behind a reverse proxy with `X-Forwarded-For` set; client IP captured for ban/rate-limit purposes (last hop; see README ops notes) — `test/ws-host.test.ts`, `test/rate-limit.test.ts`

---

## 1. Identity & Auth

### 1.1 Guest identity (`/api/identity`, `nsp_guest_id` cookie)
- [x] Fresh visitor with no cookie can land on Home (server mints guest cookie)
- [x] Guest cookie persists across reload and across tabs (same cookie ⇒ same guestId)
- [x] Clearing cookies starts a brand-new guest identity
- [x] Tampering with cookie to a non-ULID is treated as no identity (new one minted)
- [ ] Native fallbacks (`?guest=`, `?token=`) — **not tested for v1** (mobile out of scope)

### 1.2 Email/password (Better Auth)
- [x] Register with new email → signed in, redirected, session present at `/api/me`
- [x] Register with existing email → friendly error, no partial account created
- [x] Register with weak/invalid password → validation error surfaced in UI
- [x] Login with correct creds → session cookie set
- [x] Login with wrong creds → error, no session
- [x] Sign out clears session; protected routes (`POST /api/play-scripts`, `POST /api/stages`, etc.) return 401
- [~] Session survives reload (cookie-based, verified at API level); expired-session-degrades-to-guest needs a manual pass
- [x] Two browsers signed into the same account share scripts and dedupe to a single participant in any given stage — `test/ws-stage.test.ts` (§3.2 tests)

### 1.3 Guest → user upgrade
- [~] After registering/logging in mid-session, the existing `nsp_guest_id` cookie is **kept** (nothing clears or migrates it by design; verify once in a browser)
- [x] No participant rows are migrated from the guest identity to the new user identity (intentional; known papercut)
- [x] If a guest is somehow inside a stage and logs in, host status does **not** carry over (guests can never be host now, so this is unreachable)

### 1.4 Login required to host  *(implemented)*
- [x] `POST /api/stages` requires an authenticated user; guest gets 401 — `test/http-scripts.test.ts`
- [x] WS join when stage has no host **does not** assign a guest as host — `test/ws-host.test.ts`
- [x] If the only participants on a stage are guests, the lobby shows a "Waiting for a signed-in user to host" banner and start-play is disabled (server also rejects) — banner in `Lobby.tsx`; server behavior tested
- [x] Host succession only considers identities with a `userId` (guests are skipped during election)
- [x] If the host leaves and only guests remain, the stage clears the host and waits for a signed-in user

### 1.5 Auth gates (HTTP)
- [x] `POST /api/play-scripts` requires user (401 for guest)
- [x] `PUT /api/play-scripts/:id` only allowed for the script's `createdByUser`; non-owner → 404
- [x] `DELETE /api/play-scripts/:id` **admin-only** (see §2.4) — non-admin gets 403, owner non-admin also gets 403
- [x] `POST /api/play-scripts/:id/rate` requires user; guest → 401
- [~] `POST /api/livekit/token` issued only to participants currently joined to the stage, with audience receiving subscribe-only tokens (grant logic in `routes/livekit.ts`; needs a configured-LiveKit pass)

---

## 2. Scripts (CRUD, listing, ratings, tags, enabled flag)

### 2.1 Browse
- [x] Anonymous visitor sees only public **enabled** scripts
- [x] Logged-in user sees public enabled + their own enabled scripts (their disabled scripts are hidden from them too — only admins see disabled)
- [x] Admin sees all scripts including disabled, with a visible "disabled" badge (badge in `scripts/index.tsx`)
- [x] `?search=` filters case-insensitively across title/description/author
- [x] User's own ratings reflected in `userRating` on each play

### 2.2 Create / edit
- [x] Create script with cast + lines + tags (tag IDs and tag names) → persisted, retrievable
- [x] New scripts default to `enabled = true`
- [x] Edit own script — title, cast, lines, public toggle persist
- [x] Edit script you don't own → 404
- [x] Reordering / removing cast members and lines preserves index ordering
- [x] `scrollMode: 'auto'` + `autoScrollWpm` actually drive auto-scroll on stage during performance (teleprompter with progress bar, ±10s and speed controls verified in E2E screenshot of the Hamlet soliloquy)

### 2.3 Ratings  *(implemented, incl. self-rating block)*
- [x] Rate 1–5 → average and count update; user's own rating shown
- [x] Re-rate replaces previous rating, count does not double
- [x] Rate values outside 1–5 rejected with 400
- [x] Authors **cannot** rate their own scripts (server returns 403)
- [x] Guests cannot rate (server returns 401); Discussion screen shows "Sign in to rate" for unsigned users

### 2.4 Admin-only delete + `enabled` flag  *(implemented)*
- [x] DB has `enabled BOOLEAN NOT NULL DEFAULT TRUE` on `play_scripts` (migration 0006)
- [x] Admin endpoint `PATCH /api/play-scripts/:id/enabled` flips `enabled` on any script
- [x] Non-admin DELETE → 403
- [x] Admin DELETE → 200, script removed
- [x] Disabling a script: new stage creation against that `playScriptId` is rejected (400 "script unavailable")
- [x] Disabling a script: live stages currently performing it **continue uninterrupted** (stage state hydrates the play regardless of `enabled`)
- [x] Disabling a script: it disappears from `GET /api/play-scripts` for non-admins (and direct GET returns 404)
- [x] Re-enabling a disabled script restores it to listings
- [x] Admin status is derived from `ADMIN_EMAILS` env var at request time (no DB column), case-insensitive

### 2.5 Tags
- [x] `GET /api/tags` lists all tags
- [x] Creating a script with `tagNames` upserts new tags
- [x] Tag names normalized (case, trim) so duplicates merge

---

## 3. Stage lifecycle (lobby → countdown → performing → discussion)

### 3.1 Create / join
- [x] Logged-in user creates stage with no script preselected → lobby loads, code visible, they are host
- [~] Logged-in user creates stage with `?s=<scriptId>` → that play preselected (server accepts `playScriptId`; verify the UI param path)
- [x] Logged-in user cannot create a stage against a disabled script (§2.4)
- [x] Joining via 6-char code (case-insensitive) lands in the right stage
- [x] Joining a non-existent code shows "Stage not found" inline error
- [x] Joining without entering a name is blocked (client + server)
- [x] Stage code generation never produces ambiguous chars (0/O/1/I/L)
- [x] `MAX_PARTICIPANTS_PER_STAGE = 12` enforced; 13th gets "Stage is full" and is closed
- [x] Display name validated on join (see §12.4)

### 3.2 Same identity, multiple connections
- [x] Same user in two tabs: second tab displaces first; first receives `sessionReplaced` and closes; participant list still shows them once
- [x] Same guest cookie in two tabs: identical behavior to logged-in case
- [x] Reconnect after refresh reuses the existing participant row; cast roles preserved

### 3.3 Lobby actions
- [x] Click cast member → claim role; click again → release
- [x] Claiming a role another participant holds steals it (exclusive across participants)
- [x] **Same participant can hold multiple roles** (intentional doubling)
- [x] Toggling "audience" clears all of your roles
- [x] Selecting a different play clears all roles for everyone
- [x] `Ready` toggle visible to others in real time

### 3.4 Start-play preconditions  *(implemented)*
- [x] In non-director mode, `startPlay` is rejected unless **every cast role is filled**
- [x] In non-director mode, `startPlay` is rejected unless **every non-audience participant has `isReady: true`**
- [x] In director mode, only the director's `startPlay` is accepted (others get an error toast)
- [x] In director mode, the same "all roles filled" + "all non-audience ready" preconditions still apply
- [x] Lobby disables the Start button when preconditions fail and shows what's missing ("3 unfilled roles, 2 not ready") — server error message + `Lobby.tsx` status label

### 3.5 Countdown → Performing
- [x] `countdownStart` shows a 3-second countdown to all clients (verified with 3 concurrent browser contexts locally; cross-machine ~200ms timing not measured)
- [x] After countdown, `playStarted` arrives; clients render Stage with line 0 highlighted
- [x] Audience members see the same play view but no advance affordance (server rejects; UI hides the button)

### 3.6 Line advancement  *(implemented, incl. speaker-only)*
- [x] Default mode: only the **speaker of the current line** can advance (server rejects others with "Not your line")
- [x] In `directorMode`, only host advances; non-host attempts get the gated error toast
- [x] Audience cannot advance in any mode
- [x] Reaching the last line transitions to `discussion` and emits `playEnded`
- [x] Heartbeats every 5s reconcile any missed `lineAdvanced`

### 3.7 Pause / resume / back / skip / end
- [x] Host pause → broadcast system message; `nextLine` rejected with "paused" error
- [x] Host resume → flow continues from same line index
- [x] Host back: clamps at 0; non-host → "Host only"; only allowed during `performing`
- [x] Host skip: same gating as back; skipping past last line ends the play
- [~] Host end: confirmation modal mid-scene; no confirm in `discussion`; resets to lobby with everyone un-ready (server reset + un-ready tested; verify modal behavior in browser)

### 3.8 Discussion
- [x] Cast list shown with assigned actor (roles now listed next to names)
- [x] Logged-in users can rate the script from this screen
- [x] Guests see a "Sign in to rate" prompt instead of stars
- [x] Authors see their own rating disabled with a note ("Authors can't rate their own work")
- [x] "Perform again" returns to lobby preserving selected play; everyone un-ready

### 3.9 Late joiners  *(implemented)*
- [x] Joining during `countdown`, `performing`, or `discussion` automatically lands in **audience** (`isAudience: true`)
- [x] Late joiner cannot self-toggle out of audience until the stage returns to `lobby`
- [x] Late joiner sees current line index in sync immediately
- [x] System message announces the late join

---

## 4. Host & role management

### 4.1 Host election & succession
- [x] First **signed-in** identity to join a stage with no host becomes host
- [x] Guests joining a host-less stage do **not** become host (§1.4)
- [x] If host disconnects, a 15s grace window is held; system does not announce succession yet
- [x] If host returns within 15s, "host reconnected" message; no transfer
- [x] If host does not return, oldest-joined remaining **signed-in** identity is promoted; system message names them
- [x] If host is the only signed-in user and leaves, host is cleared; stage waits for a signed-in user
- [x] Two tabs: closing one tab (host's) does NOT trigger succession because the other tab is still connected

### 4.2 `hostTransferHost`
- [x] Host can transfer to any connected **signed-in** participant; old host becomes regular
- [x] Transfer to a guest is rejected ("Target must be signed in")
- [x] After transfer, host-gated actions work for new host and fail for old host

### 4.3 Reassign / unassign roles in lobby
- [x] In lobby: reassigning role moves it instantly, system message posted
- [x] Reassigning a role to someone who already holds another keeps both (doubling)
- [x] Host unassign clears the role

### 4.4 Reassign mid-performance (manual host action)
- [x] In performing: `hostReassignRole` pauses the play, sends `roleOffer` to target, posts system message
- [x] Target accepts → role moves, system message, host must manually resume
- [x] Target declines → role unchanged, play remains paused
- [x] Target disconnects before responding → offer cleared, system message, play remains paused

### 4.5 Role vacancy mid-performance (cast member leaves)  *(implemented)*
- [x] Non-director mode: when the speaker (or any role-holder) disconnects mid-play:
  - [x] Stage immediately enters `pausedByHost = true`
  - [x] System message names the vacated role
  - [x] A 45-second visible countdown starts (broadcast `roleVacated` with `claimDeadline` timestamp)
  - [x] During the countdown, any audience member can claim the vacant role with a single click; on claim, the play remains paused but the role is filled and host can resume
  - [x] Host can fast-track via existing `hostReassignRole` offer flow at any time during the countdown (bypasses timer)
  - [x] If the timer expires with no claim, the play **auto-ends** to discussion with a system message ("Play ended: role X went vacant") — note: implemented as auto-end on expiry regardless of audience availability, since an available-but-unwilling audience shouldn't hold the stage paused forever
  - [x] If the original holder reconnects within the countdown, they reclaim the role and the countdown is cancelled
- [x] Director mode: when a role vacates:
  - [x] Stage pauses
  - [x] Director uses the existing `hostReassignRole` offer flow, targeting any audience member
  - [~] Eligibility check: implemented as the offer-accept confirmation ("Accepting puts you in the cast — you can enable your camera and mic"), per the spec's allowed alternative; no server-side track check
  - [x] Target must explicitly confirm role assignment (existing `respondRoleOffer` accept path)
  - [x] No 45-second countdown in director mode
- All above: `test/ws-vacancy.test.ts`

### 4.6 Director mode
- [x] Toggle only allowed in lobby
- [x] Enabling director mode clears host's own cast assignments
- [x] Host cannot claim a role while director mode is on (error surfaced)
- [x] During performing, only director's `nextLine` advances; others get "Director controls line advance"
- [x] Disabling director mode (only in lobby) allows host to claim roles again

### 4.7 Kick / ban
- [x] Host kicks participant → target receives error, socket closed, system message; can rejoin freely
- [x] Host bans participant → target same as kick; subsequent joins from same identity blocked with "removed" message
- [x] IP ban: banned identity's IP (from `X-Forwarded-For`, last public hop) is recorded in `stage_ip_bans`; subsequent connections from that IP to this stage are rejected
- [x] Banned guest who clears their cookie: still blocked because IP ban catches them
- [x] Banned user logged out and joined as guest from a different IP: not blocked (acceptable per spec)
- [x] Host cannot kick/ban self
- [~] After kick, target's LiveKit participant is removed server-side (call wired; needs a configured-LiveKit pass)
- [x] IP bans are scoped per-stage, not global

### 4.8 Mute  *(implemented, incl. visual disable)*
- [x] Host mutes target → `forceMuted` flag set, LiveKit unpublishes target's audio (flag + admin call tested; mute now also available mid-performance via Host controls → Manage cast)
- [x] Target's local mic-toggle button is **visually disabled** with "🔇 Muted by host" (verified in E2E with LiveKit configured)
- [x] Host unmutes → button re-enabled (verified in E2E)
- [~] Target cannot re-publish audio while force-muted (UI-level: disabled toggle ignores clicks — verified; SFU-level `canPublishSources` restriction wired but needs a live A/V pass)

---

## 5. Inviting / sharing a stage

### 5.1 Invite link button
- [x] "Copy invite link" button in `Lobby.tsx` copies a working URL to clipboard
- [x] Pasting the URL in a new tab/browser lands the recipient on `/stage/:code` (E2E: fresh context navigated straight to the stage)
- [x] Toast/feedback "Invite link copied" appears on success
- [x] On clipboard API failure (e.g., insecure context), a fallback shows the URL for manual copy

### 5.2 Logged-in user clicks invite link
- [x] Lands directly on `/stage/:code`, prompted only for display name (verified in E2E)
- [x] No registration prompt
- [x] Joins as audience if stage is past lobby (§3.9), else can claim cast roles

### 5.3 Anonymous (no account) clicks invite link
- [x] Lands on `/stage/:code`, prompted for a display name only — **not** required to register
- [x] Server mints a guest cookie on first request
- [x] Joins successfully as a guest (cannot host, can claim cast roles in lobby)
- [x] A small "Sign in" affordance is visible but optional (on the join-name screen)

### 5.4 Banned identity follows a fresh invite link
- [x] Banned by identity → blocked with "removed" message
- [x] Banned by IP → blocked even if they use a fresh guest cookie (§4.7)
- [x] From a different IP, banned guest who registers a new account: **not blocked** (acceptable per spec)

### 5.5 Stage code privacy
- [x] Stage codes don't appear in any analytics/telemetry events (no analytics exist in v1)
- [x] Stage codes don't appear in server logs at INFO level (request logger masks them as `:code`; verified live)
- [x] `/api/stages/:code/exists` is rate-limited (§12.5)

---

## 6. Real-time chat  *(implemented, incl. markdown + truncation)*

### 6.1 Visibility
- [x] Public chat reaches everyone including audience
- [x] Cast-only chat reaches only non-audience participants
- [x] Toggling to/from audience changes whether you receive future cast-only messages
- [x] System messages render distinctly (no sender, italic style) — verified in E2E screenshot

### 6.2 Formatting
- [x] Bold (`**x**`), italic (`*x*` and `_x_`), inline code (`` `x` ``), strikethrough (`~~x~~`) all render (unit-tested parser + verified rendered in E2E screenshot; link markdown shown literally, bare URLs produce 0 anchor tags)
- [x] Links (`[text](url)`) render as **plain text** (never clickable)
- [x] Bare URLs (`https://example.com`) are not autolinked
- [x] Images (`![alt](url)`) render as plain text
- [x] Headings (`# x`), blockquotes (`> x`), tables, raw HTML — all left as literal text (RN `<Text>` never interprets HTML)
- [~] Emojis render correctly (native unicode; parser passes them through — verify glyphs in browser)

### 6.3 Length & truncation
- [x] `MAX_CHAT_LENGTH` const exists (256, in `@nsp/shared`)
- [x] Messages over `MAX_CHAT_LENGTH` are auto-truncated server-side before broadcast
- [x] Truncated messages get a trailing ellipsis
- [x] Client input does not hard-cap, only truncates on send

### 6.4 Persistence
- [x] Chat history is **not** persisted; rejoining users do not see past messages
- [x] System messages are also not persisted

---

## 7. Video / mic (LiveKit)

### 7.1 With LiveKit configured
- [x] **Mic and camera both off by default** on join — no auto-publish (verified in E2E against the real LiveKit room: cast joins with "Mic off / Camera off")
- [~] User can opt in to mic (toggle wired; publish starts but ICE can't complete in headless sandbox — verify on a real desktop)
- [~] User can opt in to camera (same as above — local track is created on click, publication pending)
- [x] Audience gets no publish controls; token grant is subscribe-only (control-bar absence verified in E2E; grant logic in `routes/livekit.ts`)
- [~] Long sessions (>1 hour): token TTL raised to 12h so late reconnects still work (explicit mid-session refresh not implemented — revisit if stages ever outlive 12h)
- [ ] Network drop → reconnect restores published tracks without rejoining the stage room
- [ ] Up to 12 participants render without obvious distortion / dropped frames
- [~] Force-mute (§4.8) actually unpublishes the target's audio track on the SFU (mute + `canPublishSources` restriction wired; verify live)

### 7.2 Without LiveKit configured
- [x] App shows the friendly fallback message; rest of stage features still work (full performance flow ran with LiveKit unset in E2E)
- [x] No repeated token request attempts (verified: exactly one 503 per client in the browser console — the unavoidable network-error log for that single request — and no retries)

---

## 8. Connection robustness
- [~] Server restart while clients are connected → clients show "disconnected", auto-reconnect succeeds, state intact (reconnect + participant-row reuse tested at handler level; do one real restart pass)
- [ ] WebSocket survives Wi-Fi blip on desktop (drop interface for 10s, reconnect)
- [x] Heartbeat reconciles missed `lineAdvanced`/`stageState` after suspension
- [x] Disposing the stage tab fires `onClose` which schedules host succession correctly
- [x] Sending malformed JSON → server returns "Invalid JSON" error, doesn't crash
- [x] Sending unknown message `type` → "Unknown message type" error, connection stays open
- [x] WS connection without prior identity is rejected at `join` time

---

## 9. Concurrency / edge cases
- [x] Two participants click the same cast member at the same instant → exactly one ends up holding it; UI converges (server serializes; steal semantics tested)
- [~] Host taps `Pause` then `End play` rapidly → state machine doesn't deadlock (both paths idempotent; no dedicated test)
- [x] `selectPlay` mid-lobby clears roles for everyone (no orphan role rows)
- [x] Host disconnects during a pending role offer → offer auto-cleared, play remains paused
- [~] Two simultaneous `hostKick` requests for the same target → only one succeeds, other gets graceful error ("Target not found"; no dedicated test)
- [x] Late joiner during the 45s vacancy countdown sees the countdown remaining time, not a stale value (`claimDeadline` is an absolute timestamp)
- [~] Participant rejoins after `playEnded` and discussion already started → lands in discussion with cast list visible (rejoin returns full state; no dedicated test)

---

## 10. Persistence, cleanup & lifetime  *(implemented)*

### 10.1 Persistence
- [x] Restart server → existing stage codes still resolve (until cleanup runs); participants must rejoin (all state is DB-backed)
- [x] `leftAt` is set on disconnect; `getStageState` filters by `isNull(leftAt)`
- [x] Bans persist across server restarts (both identity and IP — plain DB rows)
- [x] Rating averages recompute correctly after delete/re-rate

### 10.2 Stage cleanup
- [x] When the **last** participant leaves a stage, the stage record is deleted within 5 seconds
- [x] After cleanup, `GET /api/stages/:code/exists` returns `{exists: false}` — "as if it never existed"
- [x] An idle stage with no participants for **10 minutes** is also deleted (background reaper, 60s interval)
- [x] Cleanup also removes: `stage_participants`, `stage_participant_roles`, `stage_bans`, `stage_ip_bans` (FK cascade), in-memory connection map entries, pending offers, host-grace timers, vacancy timers
- [x] Reaper runs even if the server was restarted while the stage was idle (recovered from DB)
- [x] Cleanup does not race with a new joiner: a join during the 5s window cancels cleanup
- All above: `test/cleanup.test.ts`

---

## 11. UX / polish (desktop)
- [x] All error toasts dismissable; do not stack indefinitely — in-stage action rejections ("Not your line", host-gated actions, etc.) now route through a dedicated toast channel in `usePlayState` (separate from the sticky connection error). `ToastHost` (`components/Toasts.tsx`) renders a top-center stack that auto-dismisses after 4s, is tap-to-dismiss, and is hard-capped at 3 (`MAX_TOASTS`) so notices can never pile up. Per-screen form errors are single-slot (replaced, not stacked) by construction. **E2E-verified** (headless Chromium): fired 1 pause + 4 rejected advances in one synchronous batch → exactly 3 toasts shown (4th dropped), tap-dismiss took it to 2, and all cleared to 0 after the 4s window. Note: because every rejectable action is client-gated (buttons disable rather than send a doomed request), in practice the toast is a safety-net for server-side races — the E2E deliberately induced one to exercise it.
- [x] Disabled-button visual state distinct (already wired in `HostControls.ActionButton`)
- [~] Layout works at 1280×720 minimum (verified headless: no horizontal scroll, lobby lays out cleanly — screenshot in scratchpad) and at ≥1920×1080 (not yet checked)
- [x] Dark theme contrast meets WCAG AA on all key text — measured contrast ratios against `stage.bg` (#0a0a0a) / `stage.surface` (#141414): `text` 15.9/14.8, `gold` 9.4/8.8, `muted` 5.3/4.9, `goldDim` 5.1/4.7, `red-400` errors 7.2 — every key text token clears the 4.5:1 AA threshold for normal text (tightest is goldDim-on-surface at 4.73). Palette left unchanged.
- [x] Keyboard navigation: Tab cycles inputs/buttons sensibly — React Native Web renders `Pressable`/`TextInput` as natively focusable elements so Tab order follows source order; added an app-wide `:focus-visible` gold focus ring in `global.css` (mouse/touch focus stays ring-free). **E2E-verified**: tabbing on Home moved focus and rendered the gold ring around the focused control (screenshot). A full tab-through of every screen is still a nice human pass but the mechanism is confirmed.
- [x] Loading spinners present for any await > 200 ms (`api()` calls, WS connect) — page-level loads already used `ActivityIndicator` (scripts list/detail, stage connect); added inline spinners to every async-button path (home create/join, start-stage, rate, sign in, register) that previously showed only a "…" or text. **E2E-verified**: the create-stage button rendered the spinner + "Working…" mid-request (artificial latency added in the driver to catch it).
- [x] Empty states for: no scripts in browse ("No scripts yet."), no participants in lobby (single-participant lobby verified in E2E screenshot)
- [x] "Copy invite link" button is prominent in lobby
- [~] Force-mute disabled mic icon is unambiguous ("🔇 Muted by host" label; verify visually)

---

## 12. Security / abuse / rate limiting

### 12.1 CSRF / CORS
- [~] State-changing endpoints reject cross-origin requests without auth credentials (CORS allowlist + cookie auth; no dedicated test)
- [x] CORS allowlist matches deploy domains via `TRUSTED_ORIGINS` (no `*`)
- [x] WS upgrade respects same auth/identity as HTTP (same `resolveIdentity` path)

### 12.2 Auth endpoints
- [x] Better Auth's built-in rate limiting on sign-in / register is enabled (explicitly, in production; off in dev/test so suites aren't throttled)
- [~] Account enumeration on register is mitigated (Better Auth default behavior; not independently verified)

### 12.3 Logging hygiene
- [~] No PII in server logs (request logger logs method/path/status only; audit any error paths before launch)
- [x] No stage codes in INFO logs (masked as `:code`; verified live)

### 12.4 Display name policy  *(implemented)*
- [x] Length cap: **32 characters** (server-side; UI input `maxLength={32}`)
- [x] Profanity filter via `bad-words` (+ curated embedded-word check) — rejected names show "Please choose a different name"
- [x] Reserved tokens (case-insensitive): "Host", "Admin", "Director", "System", "Server" — rejected
- [x] Per-stage uniqueness: case-insensitive; on collision the server **auto-suffixes** `(2)`, `(3)`, etc.
- [x] Validation runs on initial join AND on rename (rename goes through the same join path)

### 12.5 Stage code lookup rate limit  *(implemented)*
- [x] `GET /api/stages/:code/exists` limited to **30 requests / minute / IP**
- [x] Overage returns 429 with `Retry-After`
- [x] Limit is per-IP, applied in middleware
- [x] Limit is documented in `README.md` for ops

### 12.6 IP capture trust
- [x] Server uses `X-Forwarded-For` last public hop for IP ban + rate limiting
- [x] If proxy is misconfigured, loud failure mode documented in README (everyone shares the proxy IP → a ban/limit hits everyone)

---

## 13. Deployment smoke
- [x] Server :3001 + Expo web :8081 both run cleanly (launched individually for the E2E pass; `pnpm dev` runs the same two commands in parallel)
- [x] `pnpm db:migrate` from a fresh SQLite file applies all migrations cleanly
- [x] `pnpm db:seed` produces the documented sample plays
- [ ] Production build of web app loads against deployed server
- [~] `EXPO_PUBLIC_API_URL` override honored (wired in `lib/config.ts`; verify against deploy)
- [ ] `/api/health` returns 200 in prod (verified locally)
- [ ] `ADMIN_EMAILS` set in prod env
- [ ] LiveKit env vars set in prod (or intentionally absent)

---

## Appendix A: Implementation Gaps

**All 13 gaps below are now implemented** (commits on `utter_refactor`, 2026-07-21) and covered by
`packages/server/test/`. Original descriptions retained for reference.

1. **Auth-required stage creation** (§1.4) — ✅ `requireAuth` on `POST /api/stages`; guest identities skipped in host election and succession; lobby banner added.
2. **Self-rating block** (§2.3) — ✅ 403 from the rate route; Discussion disables stars for authors.
3. **Admin role + `enabled` flag on plays** (§2.4) — ✅ migration 0006, `isAdminEmail()` helper, gated DELETE, `PATCH /:id/enabled`, list/get filtering, stage-creation rejection.
4. **Start-play preconditions** (§3.4) — ✅ validated in `startPlay` with a "what's missing" error; director-only start.
5. **Speaker-only line advancement** (§3.6) — ✅ `isParticipantSpeaking` check; "Not your line."
6. **Late-joiner audience lock** (§3.9) — ✅ forced `isAudience: true` outside lobby; `setAudience(false)` rejected until lobby.
7. **Mid-play role vacancy flow** (§4.5) — ✅ `roleVacated`/`roleVacancyCleared` messages, 45s claim timer, `audienceClaimRole`, reclaim-on-reconnect, host fast-track, auto-end on expiry, director-mode branch.
8. **IP ban capture** (§4.7) — ✅ `stage_ip_bans` table; IP captured at WS upgrade (XFF last hop); checked on join.
9. **Force-mute UI disable** (§4.8) — ✅ mic toggle disabled with "Muted by host" in `VideoRoom.tsx`; server also restricts publish sources.
10. **Markdown chat + length truncation** (§6.2, §6.3) — ✅ allowlist parser in `@nsp/shared`; `MAX_CHAT_LENGTH = 256` truncation in `sendChat`.
11. **Stage cleanup** (§10.2) — ✅ 5s empty-stage deletion (join cancels), 60s idle reaper (10min threshold), full in-memory teardown.
12. **Display-name policy** (§12.4) — ✅ `validateDisplayName` + `dedupeDisplayName` on join/rename.
13. **Stage code lookup rate limit** (§12.5) — ✅ 30/min/IP middleware, 429 + `Retry-After`, documented in README.

---

## Appendix B: Open design questions
- **Host vs Director consolidation.** v1 keeps them separate. Revisit for v2: collapse "host" into "director" so there is exactly one privileged user per stage and director mode becomes the only mode.
- **Guest → user upgrade migration.** v1 papercut: guest who logs in inside a stage doesn't carry their participant row. Acceptable for soft launch.
- **Banned-identity bypass.** v1 accepts that a banned guest who registers a new account from a different IP can rejoin. Per-stage bans are short-lived enough that this is fine.
- **Vacancy expiry semantics.** The plan wording ("expires with no claim and no audience available") was implemented as: expiry with no claim always auto-ends the play, even if audience members were present but didn't claim. Rationale: an unwilling audience shouldn't hold the stage paused indefinitely. Flag if the other reading was intended.
