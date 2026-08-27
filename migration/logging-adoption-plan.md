# Logging Adoption Plan — Zap (via the `slog` interface)

Status: **decided, not yet implemented.** No logging library is wired into
`migration/openelis-go/` yet — current code uses plain stdlib `log`
(`log.Println`/`log.Fatalf`) in 3 files, ~19 call sites, no structure, no
levels. This doc records the decision for whoever wires it in.

---

## Decision

| Concern              | Choice                                                          | Rationale                                                                                            |
| --------------------- | ----------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| Logging engine        | **`go.uber.org/zap`**                                             | Only mainstream Go logger with a built-in, zero-code HTTP handler for runtime level control (`AtomicLevel.ServeHTTP`) — the specific capability needed to port `LoggingController.java`'s `/logging` endpoint. |
| Public API surface     | **`log/slog`**, backed by zap via zap's `zapslog` handler adapter | Rest of the codebase writes against the stdlib-shaped `*slog.Logger` interface — zero lock-in, matches where the Go ecosystem is converging, plays well with any other library expecting an `slog.Handler`. The zap engine sits underneath, reachable directly wherever `AtomicLevel` or sampling is needed. |
| Original plan (superseded) | `tech-stack-diff.md` previously said plain `log/slog`, no third-party library | Reasonable if the only goal were structured output; underweights the runtime-level-switching requirement once `LoggingController.java` (§ below) was looked at as a concrete Go-side goal. |

**Why not plain `slog` alone:** it has the *primitive* for dynamic levels
(`slog.LevelVar`, safe to mutate concurrently) but no built-in HTTP handler —
you'd hand-write the level-parsing/HTTP-wiring code yourself, i.e. rebuild by
hand exactly the piece zap ships for free. `slog` also has no Log4j2-style
hierarchical per-package logger tree; neither does zap — that part isn't
available off-the-shelf in either, and isn't part of this decision either
way (see § What this does NOT give us, below).

**Why not zap alone (bypassing `slog`):** zap's own native API works fine
standalone, but the ecosystem is actively converging on `slog.Handler` as
the common interface (zap itself ships the adapter specifically to serve
this case). Writing application code against `*slog.Logger` instead of
`*zap.Logger` costs nothing here and avoids being locked to zap's specific
API if a future library swap is ever wanted.

---

## What this enables: `rest/logging` (currently unplanned)

`LoggingController.java`
([src/main/java/org/openelisglobal/logging/controller/LoggingController.java](../src/main/java/org/openelisglobal/logging/controller/LoggingController.java))
is admin-gated (`@PreAuthorize("hasRole('ADMIN')")`) and has three endpoints,
none of which are in `endpoint-migration-order.md` or
`endpoint-migration-taxonomy.md` today — confirmed absent from both, not
merely deferred (see the "Ops / infrastructure endpoints" note added to each
of those docs alongside this one):

1. **`GET /logging?logLevel=&logger=`** — live level change via
   `Configurator.setLevel(logger, level)`/`setRootLevel(level)`. **Go
   equivalent, now unblocked**: `zap.AtomicLevel` (or an `slog.LevelVar` if
   staying purely in the `slog` surface) exposed behind a small admin-gated
   handler. Java's single-letter level codes (`A/I/T/D/W/E/F/O`) don't need
   to be reproduced verbatim — that's a Java-ism, not a contract any real
   caller outside the admin UI depends on; map to `slog`'s
   `Debug`/`Info`/`Warn`/`Error` (a 4-level scheme, not Java's 8) unless a
   caller is found that needs the extra granularity.
2. **`GET /logging/stream`** (SSE, `text/event-stream`) — live log tail to a
   browser, backed by Java's hand-rolled `InMemoryLogAppender` (300-line ring
   buffer + pub/sub fan-out). **No off-the-shelf equivalent in zap, slog, or
   any mainstream Go logger** — Java didn't get this from Log4j2 either, it
   was custom-built there too (Log4j2 has no "stream to SSE" appender
   out of the box). Porting this means: a custom `zapcore.Core` (or
   `slog.Handler`) that also pushes formatted lines into a small ring buffer
   + subscriber set, same shape as the Java version, and a handler using
   Go's `http.ResponseWriter` flush-per-write pattern for SSE (stdlib
   `net/http` supports this directly — no library needed for the SSE
   mechanics themselves, only for hooking into the log pipeline).
3. **`GET /logging/test`** — fires one line per level. Trivial once (1) and
   (2) exist.

**Not yet decided:** whether `/logging` is worth porting at all before the
domain migration waves are further along — it's an ops convenience, not
something any real clinical/reference-data caller depends on, and it doesn't
block anything else. Recorded here so the decision is at least *possible* to
make deliberately later, rather than the endpoint staying invisible to the
whole planning process the way it was before this doc existed.

---

## What this does NOT give us

- **No Log4j2-style hierarchical logger tree.** Setting a level for
  `org.openelisglobal.provider` and having it apply to every sub-package
  underneath, with per-branch overrides, is a Log4j2/Logback feature tied to
  Java's package-name convention. Neither zap nor slog ships this. If
  fine-grained per-subsystem control is ever wanted beyond one or a handful
  of independently-switchable levels, it would need to be hand-built (e.g. a
  `map[string]*zap.AtomicLevel` keyed by component name, checked by
  longest-prefix match) — not a reason to avoid zap, since nothing else in
  the Go ecosystem ships this either; just a real gap to know about.
- **No automatic parity with Java's exact log line format.** Not a goal —
  logs aren't part of the API/DB contract this migration's parity suite
  checks (`openelis-api-e2e`), and Log4j2's specific text layout has no
  caller depending on it.

---

## Migration checklist (when this is actually implemented)

- [ ] Add `go.uber.org/zap` to `go.mod`.
- [ ] `internal/common/log` (or similar): construct one process-wide
      `*zap.Logger`, wrap via `zapslog.NewHandler(...)` into a `*slog.Logger`,
      expose that as the thing the rest of the codebase imports — mirrors
      `internal/common/db.OpenGORM()`'s role as the one place that constructs
      the shared resource.
- [ ] Replace the 3 files' plain `log.*` calls
      (`cmd/openelis/main.go`, `cmd/migrate/main.go`, `cmd/_dbsetup/main.go`)
      with the new `slog` logger.
- [ ] Resolve the dead `OE_DB_LOG=true` comment in
      [openelis-go/internal/common/db/db.go](openelis-go/internal/common/db/db.go)
      — it currently claims that env var enables GORM's SQL logger; nothing
      reads that env var anywhere. Either wire it for real (GORM accepts a
      custom `logger.Interface`; the new `slog`-backed logger could implement
      it) or remove the false claim from the comment.
- [ ] Decide (separately, see § above) whether to port `rest/logging` at all,
      and if so, add it to `endpoint-migration-order.md`'s ops-endpoints note
      as "in progress" / assign it a real wave.
