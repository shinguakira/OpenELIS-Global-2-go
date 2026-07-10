# openelis-go — minimal Go backend (migration skeleton)

Minimal Go service skeleton for the OpenELIS strangler-fig migration (see
[../OpenELIS-Go-Migration-Plan.md](../OpenELIS-Go-Migration-Plan.md)). No
endpoints are ported yet — this is just the frame that ported endpoints will be
registered into, one at a time.

## Layout

```
cmd/openelis/main.go        # entrypoint; starts the HTTP server
internal/rest/router.go     # endpoint wiring + JSON helper (health only for now)
```

## Run

```bash
cd migration/openelis-go
go run ./cmd/openelis          # listens on :8090 (override with OE_GO_ADDR)

curl -s localhost:8090/health
# {"status":"UP"}
```
