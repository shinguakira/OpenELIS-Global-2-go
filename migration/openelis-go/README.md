# openelis-go — minimal Go backend (migration)

Minimal Go service for the OpenELIS strangler-fig migration (see
[../OpenELIS-Go-Migration-Plan.md](../OpenELIS-Go-Migration-Plan.md)). Endpoints
are ported one at a time into Java-mirrored folders (`internal/<domain>/<layer>/`).

## Layout

Folders **mirror the Java source layout** (`org.openelisglobal.<domain>.<layer>`)
during migration, so Java devs navigate the same paths. Idiomatic Go
reorganization is deferred to the **end** of the migration.

Each domain follows the same 4-layer pattern:

```
internal/<domain>/valueholder/     ← entity structs  (no JSON tags, no SQL)
internal/<domain>/daoimpl/         ← SQL queries only
internal/<domain>/service/         ← business logic, calls DAO
internal/<domain>/controller/rest/ ← HTTP handler: parse → call service → DTO → JSON
```

**controller/rest MUST NOT import `database/sql` or contain SQL.**
SQL belongs exclusively in `daoimpl/`.

```
cmd/openelis/main.go               # entrypoint; wires each domain: DAO→service→controller
internal/common/web/web.go         # shared HTTP plumbing (~ org.openelisglobal.common)
```

## Ported endpoints

### a1 — Server time

| Endpoint | Java source | Plan |
|----------|-------------|------|
| `GET /rest/server-time` | `system/controller/rest/SystemRestController.java` | [a1-server-time-migration.md](../a1-server-time-migration.md) |
| `GET /health` | (Go-only liveness) | — |

### a2 — Static + first DB reads

| Endpoint | Java source |
|----------|-------------|
| `GET /rest/math-functions` | `testcalculated/controller/rest/CalculatedValueRestController.java` |
| `GET /rest/sample-item-status-types` | `common/rest/DisplayListController.java` |
| `GET /rest/supportedlocales` | `localization/controller/rest/SupportedLocaleRestController.java` |
| `GET /rest/supportedlocales/active` | same |
| `GET /rest/supportedlocales/fallback` | same |
| `GET /rest/analysis-status-types` | `common/rest/DisplayListController.java` |
| `GET /rest/sample-status-types` | same |

Plan: [a2-static-reads-migration.md](../a2-static-reads-migration.md)

### b1 — Dictionary + Test Catalog reference reads

| Endpoint | Java source |
|----------|-------------|
| `GET /rest/dictionary-categories` | `dictionary/controller/rest/DictionaryMenuRestController.java` |
| `GET /rest/uom[?type=]` | `unitofmeasure/controller/rest/UnitOfMeasureRestController.java` |
| `GET /rest/test-catalog/lab-units` | `testcatalog/controller/rest/TestCatalogEditorRestController.java` |
| `GET /rest/test-catalog/sample-types` | same |
| `GET /rest/test-catalog/panels` | same |
| `GET /rest/TestCatalog` | `testconfiguration/controller/rest/TestCatalogRestController.java` |

Plan: [b1-dictionary-testcatalog-migration.md](../b1-dictionary-testcatalog-migration.md)

## Run

```bash
cd migration/openelis-go
go run ./cmd/openelis          # listens on :8090 (override with OE_GO_ADDR)

curl -s localhost:8090/health
# {"status":"UP"}
curl -s localhost:8090/rest/server-time
# {"date":"2026-08-07","time":"20:26","timezone":"Etc/UTC"}
curl -s localhost:8090/rest/uom | jq 'length'
curl -s 'localhost:8090/rest/uom?type=SAMPLE_COLLECTION'
curl -s localhost:8090/rest/dictionary-categories | jq 'length'
curl -s localhost:8090/rest/test-catalog/lab-units
curl -s localhost:8090/rest/test-catalog/sample-types
curl -s localhost:8090/rest/test-catalog/panels
curl -s localhost:8090/rest/TestCatalog | jq '.formName, (.testSectionList | length)'
```
