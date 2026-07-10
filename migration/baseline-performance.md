# OpenELIS Global 2 — Baseline & Performance Metrics

Pre-migration baseline for `OpenELIS-Global-2-go/` (**OpenELIS Global 2**, branch
`develop`, commit `0c6a4a62d`) ahead of a planned Java → Go migration.
Companion docs: `OpenELIS-Go-Migration-Plan.md`, `OpenMRS-Analysis.md`,
`openmrs-core-go/doc/BASELINE_METRICS.md`.

> **Environment.** Measured live on a Windows 10 / **WSL2 (Ubuntu 24.04)** host,
> Docker 29.4.3, running the published `itechuw/*:develop` images (Tomcat webapp +
> PostgreSQL + HAPI FHIR + React/nginx frontend + nginx proxy). `core` + `harness`
> demo fixtures loaded, +200 generated patients, after a full E2E run — i.e. a
> **warm** instance. Runtime figures are a ceiling, not cold-idle; kept as-is for a
> fair Go comparison.

---

## 1. What this repo is

**OpenELIS Global 2** — an enterprise **Laboratory Information System (LIS)** for
public-health labs: orders → samples → results → validation → reporting, with
analyzer interop (ASTM/HL7/file), FHIR R4 interoperability, and cold-chain sample
storage.

- **Stack:** Java **21**, **Spring Framework 6.2** (traditional MVC, *not* Spring
  Boot), Jakarta EE 9 (`jakarta.*`), Hibernate/JPA, **PostgreSQL**, HAPI **FHIR
  R4**, Liquibase. Packaged as `OpenELIS-Global.war` on Tomcat. Frontend is a
  **React 17 + Carbon Design System** SPA (Vite), served separately behind nginx.
- **No Go code yet** — this is the *source* Java system for the migration.
- ~120 self-contained domain packages under `org.openelisglobal.*`, each using the
  **Valueholder → DAO → Service → Controller → Form** 5-layer pattern.
- **License:** Mozilla Public License 2.0.

---

## 2. Code & structure (measured)

| Metric | Value |
|--------|------:|
| Production Java | **362,238 LOC / 2,805 files** |
| Test Java | 90,889 LOC / 488 files |
| Frontend src (js/jsx/ts/tsx) | 201,169 LOC / 698 files |
| Domain packages (`org.openelisglobal.*`) | ~120 |
| Layer dirs (Valueholder / DAO+Impl / Service / Controller / Form) | ~120 / ~205 / ~133 / ~78 / ~51 |
| REST controllers / endpoint mappings | ~126 / ~751 |
| Liquibase changelogs / `createTable` ops | 277 / 318 |
| DB tables (`clinlims` schema, live) | **375** |
| Tracked files (git) | 7,994 |

---

## 3. Build & disk footprint

| Metric | Value |
|--------|------:|
| **`OpenELIS-Global.war`** | **219,901,654 B ≈ 210 MiB** |
| WAR exploded (deployed) | **281 MB** |
| Tomcat install + app total | 797 MB |
| **PostgreSQL `clinlims` DB on disk** | **30 MB** (with demo data) |
| Docker image — webapp (`openelis-global-2`) | **1.35 GB** |
| Docker image — FHIR (HAPI) | 2.06 GB |
| Docker image — database (Postgres + seed) | 544 MB |
| Docker image — frontend | 158 MB |
| Docker image — proxy (nginx) | 60.5 MB |
| Docker images — full stack (excl. certgen) | ≈ **4.2 GB** |

---

## 4. Runtime memory (live, warm, demo data)

`docker stats --no-stream`:

| Container | RSS |
|-----------|----------:|
| **webapp** (Tomcat + JVM, Java 21) | **1.52 GiB** |
| **fhir** (HAPI FHIR JPA server, Java 17) | **2.23 GiB** |
| database (PostgreSQL) | 116 MiB |
| frontend (nginx + static) | 10 MiB |
| proxy (nginx) | 2.6 MiB |
| **Full stack total** | **≈ 3.9 GiB** |

- Webapp JVM process RSS ≈ **1,615,764 KB (~1.54 GiB)**; OpenJDK 21 on Tomcat,
  JDBC to `postgresql://db.openelis.org:5432/clinlims`.
- **The FHIR server (2.2 GiB) is a *separate* HAPI JPA app** — larger than the
  webapp itself. If the Go port keeps HAPI as a facade, that footprint stays; only
  the **webapp (1.5 GiB)** is the "core" being replaced.

---

## 5. Performance (measured live)

### 5.1 Startup time (cold container → ready)
| Service | Time | Detail |
|---------|-----:|--------|
| **Webapp (Tomcat)** | **~129 s** | `Server startup in 128,887 ms`; WAR deploy alone **49.6 s** (Spring context + Hibernate + Liquibase) |
| **FHIR (HAPI)** | **~72 s** | `Started Application in 55.3 s`; Tomcat `startup in 72,334 ms` |

A ~2-minute boot is a Spring/Hibernate/servlet-container cost — the headline
contrast for a Go rewrite (native binary starts in **milliseconds**).

### 5.2 Endpoint latency (warm, 5-sample avg)
| Endpoint | Avg | HTTP | Note |
|----------|----:|-----:|------|
| `GET /` (React shell) | **7 ms** | 200 | nginx static — fast |
| `GET /api/OpenELIS-Global/rest/menu` | **14 ms** | 302 | app round-trip; 302 = redirect to login (no session in probe) |
| `GET :8444/fhir/metadata` | — | 000 | not captured this run (TLS/port) — re-measure via `:8081` http |

App responsiveness is fine once warm; the cost is **startup + memory**, not
per-request latency at idle load. (No sustained-load throughput test was run —
add a k6/JMeter pass against `/rest/*` + FHIR for a rps/peak-RSS figure like the
OpenMRS baseline's 4,000 rps / 1,415 MB.)

### 5.3 Stability
- All containers `restarts=0` at measurement (webapp, fhir, database).
- **Op note:** the `frontend` container ships with `restart: no` and a `tty:true`
  that makes its nginx exit ~38 s after `up`, which in turn crash-loops the proxy.
  Workaround applied here: `docker update --restart unless-stopped
  openelisglobal-front-end openelisglobal-proxy`. Worth fixing in the compose file.

---

## 6. Java → Go headline gap to track

| | Java baseline | Go target |
|---|---:|---|
| Core artifact | 210 MiB WAR + JVM + Tomcat | single ~10–30 MB binary |
| **Cold start** | **~129 s** (webapp) | ~milliseconds |
| Webapp RSS | 1.52 GiB | (measure later) |
| Full-stack RSS | ~3.9 GiB (incl. 2.2 GiB HAPI FHIR) | (measure later) |
| Production Java LOC | 362,238 | (measure later) |
| DB tables | 375 | 375 (schema kept) |
| Warm request latency | 7–14 ms | (should match or beat) |

---

## 7. Top migration risks (see `OpenELIS-Go-Migration-Plan.md`)

1. **Regulated lab data** — result-validation & reference-range rules
   (`resultvalidation`, `resultlimits`) are clinical-safety logic; build a JSON/DB
   parity harness. Any behavioral drift is a patient-safety risk.
2. **FHIR R4 fidelity** — HAPI is a 2.2 GiB subsystem; facade it or use a Go FHIR
   library, do **not** hand-reimplement (plan D5).
3. **Analyzer interop** — ASTM/HL7/file + the `openelisglobal-plugins` Java plugin
   model can't load in Go (D2/D6). Decide "replace core" vs "permanent hybrid".
4. **Hibernate lazy-load / cascade** — no ORM in Go; the object graph each service
   touches becomes explicit SQL.
5. **Keep the PostgreSQL schema** (Liquibase-owned, 375 tables) so Go and Java can
   share the DB during a strangler-fig cutover.
6. **Audit fields / `@Transactional`** — set invisibly by Hibernate/Spring; must be
   centralised in Go or every write is subtly wrong.

---

## 8. The case for Go — measured, with OpenELIS-specific meaning

Not abstract "Go is fast" claims — each point is tied to a number from *this*
system and to what OpenELIS actually is: a LIS run in **resource-constrained
public-health labs**, shipped via an **offline installer**, holding **regulated
patient data**, and maintaining **many concurrent analyzer connections**.

| # | Proof point (measured here) | Why it matters for OpenELIS | Go outlook |
|---|-----------------------------|------------------------------|-----------|
| 1 | **~3.9 GiB** full-stack RSS (1.5 GiB webapp + 2.2 GiB FHIR); JVM has **no `-Xmx`** → default cap = ¼ host RAM | Target sites are low-resource labs. 4 GiB RAM just to boot dictates real hardware cost and blocks small/edge boxes. | A Go core typically runs in **tens of MB** — fits a mini-PC / small VPS / edge device. Memory ≈ working set, not ¼ of RAM. |
| 2 | **~129 s** webapp cold start (WAR deploy 49.6 s) | Slow boot = long recovery after crash/deploy, painful **HA failover** (OpenELIS even documents a failover setup), no scale-to-zero. | Native binary starts in **milliseconds** → fast failover, rolling deploys, autoscale, scale-to-zero. |
| 3 | **210 MiB WAR + JDK 21 + Tomcat + 247 jars (175 MB lib)** | The **offline installer** must bundle a JVM, a servlet container, and hundreds of jars — big download, many field failure modes, JVM/Tomcat patching in the field. | **One static binary** (~10–30 MB). No JVM or Tomcat to install/patch. Installer collapses to a single file. |
| 4 | **247 dependency jars** on the classpath (534 jars across Tomcat) | Regulated patient data ⇒ security-sensitive. Each jar is a CVE/supply-chain liability (Log4Shell was *one* jar); runtime classpath enables reflection/deserialization gadget chains. | Smaller, **statically-analyzable** dependency tree (`govulncheck`), no runtime classpath, no reflection gadget surface. |
| 5 | **89 live JVM threads at idle** (thread-per-request / pool model) | The analyzer bridge holds **many concurrent instrument connections** (ASTM sockets, HL7, file pollers). Threads cost ~1 MB stack each and cap concurrency. | **Goroutines** (~KBs each) handle thousands of concurrent connections cheaply — a direct fit for analyzer interop. |
| 6 | Warm latency 7–14 ms, but JVM pays **JIT warm-up + G1 GC pauses** | First requests after each ~129 s restart are slow; GC pauses jitter p99. | No JIT warm-up, low-latency GC → consistent p99 from request one. |
| 7 | Images **1.35 GB (webapp) + 2.06 GB (FHIR)** | Slow pulls, slow scaling, heavy registry/edge sync — costly for a **Consolidated Server** fanning out to many facilities. | `scratch`/distroless images **~10–30 MB** pull in seconds; cheap to scale and sync to the edge. |
| 8 | Hibernate lazy-load, Spring AOP `@Transactional`/`@Authorized` are **invisible** in source | The plan's §4 risks (silent data-fetch, transaction, authz bugs) hide inside framework magic. | Go makes fetch/transaction/authz **explicit code** — reviewable, testable, no hidden behavior. |

**Density math (operational cost).** At ~4 GiB/instance you fit roughly **one**
OpenELIS per 4 GiB node; a Go core at ~50 MB fits **dozens** on the same box —
directly relevant to the multi-facility Consolidated Server.

### Honest counterweights (so this stays analysis, not a sales pitch)
- **FHIR (2.2 GiB HAPI) and the Java analyzer plugins are the hard parts.** A full
  Java retirement may be a **permanent hybrid** (plan D2/D5) — the wins above apply
  to the *core*, not necessarily to FHIR/plugins on day one.
- **Rewrite risk on regulated clinical logic** — parity harness is mandatory;
  behavioral drift is a patient-safety issue, not a bug.
- **Ecosystem maturity** — Java's FHIR/health libraries are mature; Go equivalents
  are thinner. Weigh build-vs-reuse per subsystem.
- **Team skills & timeline** — a 362 KLOC / 375-table system is a multi-quarter
  effort; the gains compound only after the P0 template lands (plan §3).

**One-line takeaway:** the Java baseline's cost is **startup + memory + artifact
weight** (129 s, 3.9 GiB, 210 MiB WAR, 247 jars), *not* idle latency — and those
three are exactly what a Go rewrite structurally removes, which maps cleanly onto
OpenELIS's low-resource, offline, edge-deployed reality.

---

## Appendix — measurement method
- LOC/files: `find … -name '*.java' | xargs cat | wc -l` (prod = `src/main/java`).
- WAR: `docker exec openelisglobal-webapp ls -la …/OpenELIS-Global.war`.
- Memory: `docker stats --no-stream` + in-container `ps -o rss` for the JVM.
- Startup: `docker logs …` → `Server startup in … ms` / `Deployment … finished in`.
- Latency: `curl -sk -o /dev/null -w '%{time_total}'`, 5-sample average.
- DB size: `SELECT pg_size_pretty(pg_database_size('clinlims'))`.
- Images/tables: `docker images`; `information_schema.tables` (`table_schema='clinlims'`).
- Host: Windows 10 + WSL2 Ubuntu 24.04, Docker 29.4.3. Commit `0c6a4a62d` (`develop`).
