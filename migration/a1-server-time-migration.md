# a1 — First Sample Migration: `GET /rest/server-time`

Status: **plan → implement**
Scope: **one endpoint.** The pilot that proves the whole strangler-fig mechanism
end to end. Branch: `migration/a1-pilot-server-time` (forked from
`migration-base`). Type **A** (static/computed). See
[endpoint-migration-taxonomy.md](endpoint-migration-taxonomy.md) and
[branch-naming.md](branch-naming.md).

**Definition of done:** the Go port serves `GET /rest/server-time` with a
response the parity e2e accepts as equivalent to the Java baseline.

---

## 1. Scope (and non-scope)

| In scope | Out of scope (later waves) |
|----------|----------------------------|
| `GET /rest/server-time` handler in Go | Any other endpoint |
| Response shape parity with Java | nginx route flip / auth pass-through (§7 Level 2) |
| Local parity test against the Go service | DB, ORM, FHIR — this endpoint touches none |
| IANA timezone resolution (the one real risk) | Session auth in Go (proxy/Java keeps owning it) |

This endpoint has **no DB, no auth logic of its own, no dependencies** — which is
exactly why it is the pilot.

---

## 2. The Java source (the contract to reproduce)

`src/main/java/org/openelisglobal/system/controller/rest/SystemRestController.java`

```java
@RestController
@RequestMapping("/rest")
public class SystemRestController {

    @GetMapping(value = "/server-time", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<Map<String, Object>> getServerTime() {
        try {
            Map<String, Object> response = new HashMap<>();
            ZoneId zoneId = ZoneId.systemDefault();
            LocalDate now = LocalDate.now(zoneId);
            LocalTime time = LocalTime.now(zoneId);
            response.put("date", now.format(DateTimeFormatter.ISO_LOCAL_DATE));
            response.put("time", time.format(DateTimeFormatter.ofPattern("HH:mm")));
            response.put("timezone", zoneId.getId());
            return ResponseEntity.ok(response);
        } catch (Exception e) {
            LogEvent.logError(...);
            return ResponseEntity.internalServerError().build();
        }
    }
}
```

### Behavior contract

| Aspect | Java behavior |
|--------|---------------|
| Method / path | `GET /api/OpenELIS-Global/rest/server-time` |
| Content-Type | `application/json` |
| Auth | **Required.** Under `/rest/**` → `authenticated()` in `SecurityConfig` (not in the permitAll OPEN_PAGES list). |
| CSRF | N/A for GET |
| Success body | `{ "date": "<yyyy-MM-dd>", "time": "<HH:mm>", "timezone": "<IANA zone id>" }` |
| `date` | `LocalDate.now(systemDefault)` → `ISO_LOCAL_DATE` (`2026-07-10`) |
| `time` | `LocalTime.now(systemDefault)` → `HH:mm` (24h, `20:26`) |
| `timezone` | `ZoneId.systemDefault().getId()` — IANA id, e.g. `Etc/UTC` (container TZ) |
| Error | 500 empty body on exception (not reproduced — no failure path in Go) |

JSON **key order is not part of the contract** (Java uses `HashMap`; Jackson
emits hash order). Parity compares the *parsed object*, not the raw string.

---

## 3. Java → Go code mapping

Folder paths **mirror the Java source** (`system/controller/rest/`); the Go
package for the file is `rest` (the leaf dir), imported as `systemrest`.

| # | Java | Go |
|---|------|-----|
| M1 | `@RestController @RequestMapping("/rest")` class (`system/controller/rest/`) | `internal/system/controller/rest/system.go`, `package rest` |
| M2 | `@GetMapping("/server-time")` | `web.Register(mux, "GET", "rest/server-time", ServerTime)` in `Routes(mux)` |
| M3 | `Map<String,Object> response = new HashMap<>()` | `map[string]string{...}` |
| M4 | `ZoneId.systemDefault()` | `systemZoneID()` (TZ env → IANA, §5) |
| M5 | `LocalDate.now(zone).format(ISO_LOCAL_DATE)` | `time.Now().Format("2006-01-02")` |
| M6 | `LocalTime.now(zone).format("HH:mm")` | `time.Now().Format("15:04")` |
| M7 | `zoneId.getId()` | `systemZoneID()` returns the IANA id |
| M8 | `ResponseEntity.ok(map)` (Jackson → JSON) | `web.WriteJSON(w, 200, map)` (`encoding/json`) |
| M9 | `produces = APPLICATION_JSON_VALUE` | `web.WriteJSON` sets `Content-Type: application/json` |
| M10 | Spring Security `/rest/**` authenticated | **not** in Go — auth stays at proxy/Java (§7) |
| M11 | `try/catch → 500` | omitted — no failure path; handler cannot error |

Shared plumbing (`web.Register`, `web.WriteJSON`) lives in
`internal/common/web/` — the analog of `org.openelisglobal.common`.

### Target Go code

`internal/system/controller/rest/system.go`
```go
// Package rest ports org.openelisglobal.system.controller.rest.
package rest

import (
	"net/http"
	"os"
	"time"

	"openelis-go/internal/common/web"
)

// Routes registers the system REST endpoints (mirrors SystemRestController).
func Routes(mux *http.ServeMux) {
	web.Register(mux, "GET", "rest/server-time", ServerTime)
}

// ServerTime reproduces Java SystemRestController#getServerTime:
//   GET /rest/server-time -> {"date","time","timezone"}
func ServerTime(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	web.WriteJSON(w, http.StatusOK, map[string]string{
		"date":     now.Format("2006-01-02"), // ISO_LOCAL_DATE
		"time":     now.Format("15:04"),      // HH:mm (24h)
		"timezone": systemZoneID(),           // IANA zone id
	})
}

// systemZoneID mirrors ZoneId.systemDefault().getId() — an IANA id
// (e.g. "Etc/UTC"), NOT Go's zone abbreviation ("UTC"/"JST").
func systemZoneID() string {
	if tz := os.Getenv("TZ"); tz != "" {
		return tz // container sets TZ=Etc/UTC → matches Java
	}
	if name := time.Local.String(); name != "" && name != "Local" {
		return name // IANA name when Go resolved one
	}
	name, _ := time.Now().Zone() // last-resort abbreviation
	return name
}
```

`cmd/openelis/main.go` — wire the domain's routes (mirrors Spring
auto-discovering `@RestController` beans; one line per ported domain):
```go
systemrest "openelis-go/internal/system/controller/rest"
...
systemrest.Routes(mux) // a1
```

---

## 4. Field-by-field format equivalence

| Field | Java format | Go layout | Same? |
|-------|-------------|-----------|-------|
| date | `ISO_LOCAL_DATE` = `yyyy-MM-dd` | `2006-01-02` | ✅ identical |
| time | `HH:mm` (24-hour) | `15:04` | ✅ identical |
| timezone | `ZoneId.getId()` IANA | `systemZoneID()` | ⚠️ see §5 |

`date`/`time` are clock-dependent, so parity asserts **format/shape**, not exact
value (both read the same wall clock in the same zone within a second).

---

## 5. The one divergence risk: `timezone`

- Java `ZoneId.systemDefault().getId()` → a full **IANA id** (`Etc/UTC`,
  `Asia/Tokyo`).
- Go `time.Now().Zone()` → an **abbreviation** (`UTC`, `JST`). **Not equal.**

**Resolution (implemented in `systemZoneID`):** prefer the `TZ` env var. In the
container `TZ=Etc/UTC` is set (verified in
[tech-stack-diff.md](tech-stack-diff.md) §5), so Go returns `Etc/UTC` — matching
Java. On a dev box with `TZ` unset, `time.Local.String()` yields the IANA name on
Linux/macOS; Windows falls back to the abbreviation (a dev-only mismatch, not a
container/production one).

Parity is asserted **in the same environment** (same `TZ`) so both sides read the
same zone id.

---

## 6. Files touched

Folder layout **mirrors the Java source** (`system/controller/rest/…`) during
migration; idiomatic Go reorg is deferred to the end.

| File | Change |
|------|--------|
| `openelis-go/internal/system/controller/rest/system.go` | **new** — `ServerTime`, `Routes`, `systemZoneID` (ports `SystemRestController`) |
| `openelis-go/internal/common/web/web.go` | shared `WriteJSON` + `Register` helper (~ `org.openelisglobal.common`) |
| `openelis-go/cmd/openelis/main.go` | wire `systemrest.Routes(mux)` + `/health` |
| `openelis-go/README.md` | note the ported endpoint + Java↔Go path map |

No new dependencies. `go build`/`go vet` stay clean.

---

## 7. e2e / parity gate — how we prove it passes

The e2e we already wrote is
`openelis-api-e2e/tests/readonly/03-type-a.spec.ts`, whose **a1 pilot** case
asserts: `GET rest/server-time` → **200** and an authed (non-login-HTML) JSON
body.

**Level 1 — port correctness (this task).** Run the a1 assertions against the Go
service directly (`http://localhost:8090/`). server-time needs no auth in Go. The
`03-type-a` spec has **two** a1 cases:

1. `returns 200 and an authed JSON body` — status + not-login-HTML.
2. `shape + IANA timezone (Java/Go parity)` — asserts `Content-Type:
   application/json`, keys `{date,time,timezone}`, `date`=`yyyy-MM-dd`,
   `time`=`HH:mm`, and — the timezone-compatibility guarantee — that `timezone`
   is a **valid IANA id** (`Etc/UTC`, `Asia/Tokyo`, `UTC`, `GMT`), **not** a Go
   abbreviation like `JST`. This is the check that would fail loudly if the Go
   port emitted an abbreviation (see §5).

> **Environment note:** the IANA-timezone assertion holds wherever `time.Local`
> resolves an IANA zone — the Linux container (`TZ=Etc/UTC`) and Linux CI. On a
> **Windows dev box with `TZ` unset**, the current Go falls back to the
> abbreviation (`JST`) and this assertion RED-flags it — which is correct: it is
> exactly the incompatibility the test exists to catch. Run the Go process with
> `TZ` set (as the container does) for a faithful parity run.

**Level 2 — full strangler integration (later, NOT this task).** Flip the nginx
route `/rest/server-time` → Go while auth/session stays on Java, then run the
full authed suite through the proxy (`https://localhost/...`). Documented here so
the boundary is explicit; it is not required to call the port itself correct.

> Per branch policy, the e2e **spec** is not modified here (it already exists).
> Level-1 execution targets the Go service via config/env only — no new
> committed e2e artifacts without asking.

---

## 8. Checklist

- [ ] `system.go` created (handler + `systemZoneID`)
- [ ] route registered in `router.go`
- [ ] `go build ./...` and `go vet ./...` clean
- [ ] Go service serves `GET /rest/server-time` → `{date,time,timezone}`
- [ ] Level-1 parity assertions pass against the Go service
- [ ] response keys match the Java baseline; `timezone` matches under same `TZ`
- [ ] README updated
