# openelis-go — minimal Go backend (migration)

Minimal Go service for the OpenELIS strangler-fig migration (see
[../OpenELIS-Go-Migration-Plan.md](../OpenELIS-Go-Migration-Plan.md)). Endpoints
are ported one at a time into Java-mirrored folders (`internal/<domain>/controller/rest/`).

## Layout

Folders **mirror the Java source layout** (`org.openelisglobal.<domain>.<layer>`)
during migration, so Java devs navigate the same paths. Idiomatic Go
reorganization is deferred to the **end** of the migration.

```
cmd/openelis/main.go                       # entrypoint; wires each domain's routes
internal/common/web/web.go                 # shared HTTP plumbing (~ org.openelisglobal.common)
internal/system/controller/rest/system.go  # SystemRestController port (server-time — a1)
```

| Go path | Java path |
|---------|-----------|
| `internal/system/controller/rest/system.go` | `system/controller/rest/SystemRestController.java` |
| `internal/common/web/` | `org.openelisglobal.common` (shared) |

## Ported endpoints

| Endpoint | Type | Java source | Plan |
|----------|------|-------------|------|
| `GET rest/server-time` | A (a1) | `system/controller/rest/SystemRestController.java` | [a1-server-time-migration.md](../a1-server-time-migration.md) |
| `GET /health` | — | (Go-only liveness) | — |

## Run

```bash
cd migration/openelis-go
go run ./cmd/openelis          # listens on :8090 (override with OE_GO_ADDR)

curl -s localhost:8090/health
# {"status":"UP"}
curl -s localhost:8090/rest/server-time
# {"date":"2026-07-10","time":"20:26","timezone":"Etc/UTC"}
```
