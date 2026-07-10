# OpenELIS → Go — Branch Naming

Status: **draft / proposal**
Companion to [endpoint-migration-order.md](endpoint-migration-order.md) and
[endpoint-migration-taxonomy.md](endpoint-migration-taxonomy.md).

Every migration branch **forks from `migration-base`** (never from `develop`).

Convention: `migration/<type><seq>-<slug>` — lowercase kebab. `<type>` is the
taxonomy letter (A–J), `<seq>` orders branches within a type, `<slug>` is a
short human label.

---

## Type A branches (two)

| Branch | Scope | Endpoints |
|--------|-------|-----------|
| `migration/a1-pilot-server-time` | **First sample migration** — one endpoint ported end-to-end (Go handler + nginx route + parity test) to prove the whole mechanism and rollback. | `GET rest/server-time` |
| `migration/a2-static-reads` | **All of Type A** — the remaining static / computed / read-only config endpoints. | `configuration-properties`, `open-configuration-properties`, `math-functions`, `analysis-status-types`, `sample-status-types`, `sample-item-status-types`, `supportedlocales/*`, `menu`, `menu/{elementId}`, `admin/menu/{elementId}` |

Both fork from `migration-base`. `a1` is the pilot; `a2` covers Type A in full.

---

## Remaining types (one branch each, forked from `migration-base`)

| Branch | Type / wave |
|--------|-------------|
| `migration/b1-dictionary-testcatalog` | B — reference reads (Wave 1) |
| `migration/b2-org-provider` | B — org + provider reads (Wave 2) |
| `migration/c1-patient-reads` | C/D — patient reads (Wave 3) |
| `migration/c2-sample-order-reads` | C/D — sample/order reads (Wave 4) |
| `migration/c3-result-reads` | C — result reads (Wave 5) 🔴 |
| `migration/e1-config-crud` | E — admin CRUD write-proving ground (Wave 6) |
| `migration/e2-testcatalog-writes` | E — test-catalog writes (Wave 6) |
| `migration/f1-clinical-writes` | F — order→result→validate (Wave 7) 🔴 |
| `migration/h-<module>` | H — one per feature module, e.g. `migration/h-inventory`, `migration/h-eqa`, `migration/h-shipping` (Wave 8) |
| `migration/i1-fhir-interop` | I — e-orders + FHIR facade (Wave 9) |

Type G (binary/file) rides on its owning context's branch. Type J
(analyzer/plugin) gets no migration branch in v1 — it stays on Java.

---

## Rules

- Fork point: always `migration-base`.
- One branch = one reviewable unit (a type, a wave slice, or a single H module).
- Merge target: `migration-base` (not `develop`) during coexistence.
- `a1-pilot-server-time` merges first; it establishes the Go service, proxy
  wiring, and parity harness the later branches build on.
