# OpenELIS → Go — Branch Naming

Status: **draft / proposal**
Companion to [endpoint-migration-order.md](endpoint-migration-order.md) and
[endpoint-migration-taxonomy.md](endpoint-migration-taxonomy.md).

There are **two tracks**, and they fork from **different** base branches:

| Track | Purpose | Fork from | Prefix |
|-------|---------|-----------|--------|
| **Migration** | Go port implementation, one endpoint/type at a time | `migration-base` | `migration/` |
| **e2e** | e2e / parity **test** additions or updates | `develop` | `e2e` |

> **Migration** branches **never** fork from `develop`. **e2e** branches
> **always** fork from `develop` (not `migration-base`), because the test suite
> follows the mainline lineage.
>
> **Before adding or updating any e2e test, ASK the user whether it should be
> added to e2e.** Do not add e2e specs unprompted.

Convention (migration): `migration/<type><seq>-<slug>` — lowercase kebab.
`<type>` is the taxonomy letter (A–J), `<seq>` orders branches within a type,
`<slug>` is a short human label.

---

## Type A branches (two)

| Branch | Scope | Endpoints |
|--------|-------|-----------|
| `migration/a1-pilot-server-time` | **First sample migration** — one endpoint ported end-to-end (Go handler + nginx route + parity test) to prove the whole mechanism and rollback. | `GET rest/server-time` |
| `migration/a2-static-reads` | **Static + first single-table DB reads + status-type reference data.** 7 endpoints; Stages 1–3 complete, Stage 4 (i18n `display_key` lookup) planned. See [a2-static-reads-migration.md](a2-static-reads-migration.md). | `math-functions`, `sample-item-status-types`, `supportedlocales`, `supportedlocales/active`, `supportedlocales/fallback`, `analysis-status-types`, `sample-status-types` |

Both fork from `migration-base`. `a1` is the pilot; `a2` covers the safe static
and first single-table DB reads including i18n infrastructure for all future units.

Deferred from original a2 list — moved to their own branches:

| Endpoint | Goes to |
|----------|---------|
| `open-configuration-properties` | a dedicated config branch |
| `configuration-properties` | a dedicated config branch |
| `menu`, `menu/{elementId}`, `admin/menu/{elementId}` | Type C (`migration/c-menu`) |

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

## e2e test branches

e2e / parity **test** work (the `openelis-api-e2e` suite, `frontend/playwright`,
new parity specs) is a **separate track**:

- Fork point: **`develop`** (not `migration-base`).
- Prefix: **`e2e`** — e.g. `e2e/type-a-parity`, `e2e/config-crud`.
- Merge target: `develop`.
- **Ask the user first** whether the change should be added to e2e before
  creating the branch or adding specs.

---

## Rules

- **Migration** fork point: always `migration-base`. **e2e** fork point: always
  `develop`.
- One branch = one reviewable unit (a type, a wave slice, a single H module, or
  one e2e change set).
- Merge target: migration → `migration-base`; e2e → `develop`.
- `a1-pilot-server-time` merges first; it establishes the Go service, proxy
  wiring, and parity harness the later branches build on.
