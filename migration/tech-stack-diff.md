# OpenELIS — Tech Stack Diff (Java → Go)

Status: **draft / reference** Companion to
[OpenELIS-Go-Migration-Plan.md](OpenELIS-Go-Migration-Plan.md) and
[baseline-performance.md](baseline-performance.md). Versions below are read from
the actual `pom.xml` / `frontend/package.json` on `migration-base`.

This doc is a side-by-side of the **current Java stack** and the **proposed Go
target**, so the substitution for each layer is explicit before any porting.

---

## 1. Platform / runtime

| Concern     | Java (current)                                                                    | Go (target)                                           | Migration note                                                 |
| ----------- | --------------------------------------------------------------------------------- | ----------------------------------------------------- | -------------------------------------------------------------- |
| Language    | **Java 21.0.11 LTS** (Temurin / Eclipse Adoptium)                                 | **Go 1.26**                                           | GC both; Go compiles to a single static binary.                |
| Runtime     | **Apache Tomcat 10.1.57** servlet container, WAR deploy                           | stdlib `net/http` server, standalone binary           | No servlet container, no WAR. `go build` → one executable.     |
| EE platform | `jakarta.*` namespace: servlet-api **6.0.0**, JSP 4.0.0, JSTL 3.0.2 (Servlet 6.0) | none (no JSP/servlet model)                           | Legacy JSP/MVC layer is dropped, not ported (React is the UI). |
| Packaging   | Maven → `.war` (~210 MiB) + 247 dependency jars                                   | `go build` → ~10–20 MiB static binary, 0 runtime jars | See [baseline-performance.md](baseline-performance.md).        |
| Startup     | JVM warm-up + WAR deploy (~120s cold)                                             | native binary (sub-second)                            | Big operational win; the core reason for the migration.        |
| Memory      | webapp RSS ~1.5 GiB, FHIR ~2.2 GiB                                                | expected order-of-magnitude lower                     | To be measured per ported context.                             |

---

## 2. Primary libraries

| Role             | Java lib + version                                                                  | Go equivalent (proposed)                                                                                                                       | Note                                                                  |
| ---------------- | ----------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| Web / routing    | **Spring Framework 6.2.17** MVC (traditional, **not** Boot)                         | stdlib `net/http` (`ServeMux`, method routing); `chi`/`echo` only if needed                                                                    | Controllers → handlers. No Spring context.                            |
| Security / auth  | **Spring Security 6.2.8**                                                           | custom middleware + `context.Context` principal                                                                                                | Port role/rolemodule privilege checks explicitly.                     |
| DI / wiring      | Spring IoC container                                                                | plain constructor wiring in `main`                                                                                                             | No annotations; wire dependencies by hand.                            |
| ORM              | **Hibernate 5.6.15.Final** (+ validator 8.0.2, search 6.1.8)                        | **GORM** (`gorm.io/gorm` + `gorm.io/driver/postgres`) — adopted starting b1, decision + rationale in [orm-adoption-plan.md](orm-adoption-plan.md) | **Superseded plan.** Originally proposed as raw `pgx`+`sqlc`; switched to GORM (de facto Go standard, ~37k stars) before b1 shipped — see the plan doc for why. Still explicit-query style (`db.Raw().Scan()` everywhere, not GORM's lazy-ish `db.Find()`), so the "no Hibernate-style lazy loading" shift below is unaffected. |
| DB driver        | **PostgreSQL JDBC 42.7.11**                                                         | `jackc/pgx` v5 (used directly in `cmd/loadbaseline`'s `COPY` path; everywhere else it's underneath GORM's postgres driver, not called directly) | Same wire protocol.                                                   |
| Database         | **PostgreSQL 14+**                                                                  | **PostgreSQL 14+ (unchanged)**                                                                                                                 | Schema kept as-is; both apps share it during coexistence.             |
| Schema migration | **Liquibase 4.8.0** (993 changesets / 277 files)                                    | **Liquibase kept** during coexistence; later `goose` (extraction + tooling done, see [liquibase-to-goose-plan.md](liquibase-to-goose-plan.md)) | Liquibase stays the single schema owner until Java retires.           |
| FHIR             | **HAPI FHIR 7.0.2** (structures-r4, server, client, base) + `org.hl7.fhir.r4` 6.9.4 | **facade to HAPI server** (kept) or a Go FHIR lib                                                                                              | Do **not** hand-roll FHIR R4 (plan D5).                               |
| HL7 v2           | **HAPI HL7v2 2.5.1**                                                                | native Go **or** stays Java                                                                                                                    | Analyzer interop — likely permanent hybrid (plan D2/D6).              |
| JSON             | **Jackson 2.18.6** (+ hibernate5, jdk8, jsr310, csv)                                | stdlib `encoding/json`                                                                                                                         | Match field names/shapes exactly for parity.                          |
| Reports / PDF    | **JasperReports 6.15.0**                                                            | **stays Java** initially; Go PDF lib later                                                                                                     | PDF byte-parity is hard; assert on data, not bytes (taxonomy Type G). |
| XML binding      | **Castor 1.4.1**                                                                    | stdlib `encoding/xml` (only where still needed)                                                                                                | Mostly legacy; port only if the path survives.                        |
| Logging          | **Log4j 2.17.1**                                                                    | stdlib `log/slog`                                                                                                                              | Structured logging.                                                   |
| Build            | **Maven**                                                                           | `go build` / `go test`                                                                                                                         | No plugin/lifecycle machinery.                                        |
| Unit test        | JUnit + **Testcontainers 1.19.8**                                                   | stdlib `testing` + table tests; Testcontainers-go if needed                                                                                    | Port high-value rule tests (validation, ranges).                      |
| Parity / e2e     | Playwright 1.58 (API + UI)                                                          | **same suite, unchanged**                                                                                                                      | The language-neutral parity oracle (`openelis-api-e2e`).              |

---

## 3. Frontend (unchanged — not part of the port)

| Concern       | Version                                 | Note                                                           |
| ------------- | --------------------------------------- | -------------------------------------------------------------- |
| Framework     | **React 17.0.2**                        | Stays. The Go backend serves the same REST + FHIR it consumes. |
| Design system | **@carbon/react 1.15**                  | Unchanged.                                                     |
| i18n          | **react-intl 5.20**, **i18next 21.10**  | New keys still land in `en.json` only.                         |
| Build / test  | **Vite 8**, **Vitest 4**, Node 20/22/24 | Unchanged.                                                     |

The frontend is the **contract consumer**, not a migration target. It keeps
talking to the same endpoints; only what serves them changes.

---

## 4. The three shifts that matter most

1. **Hibernate ORM → GORM, explicit-query style.** Java walks object graphs
   lazily (`sample.getSampleItems().getAnalyses()...`, fetched on access); Go
   loads exactly what each query specifies (`db.Raw(...).Scan(&result)` — see
   [orm-adoption-plan.md](orm-adoption-plan.md)). This is where behavior
   silently diverges — every service method's touched object graph must be
   mapped deliberately, and where Java's lazy-loading timing itself turns out
   to matter (e.g. a `lazy="true"` collection serializing as `null` once the
   session closes — a real, confirmed case in
   [b2-org-provider-migration.md](b2-org-provider-migration.md) §3.2 #8).
   (See plan §4.)
2. **Spring context → plain wiring.** No IoC/DI, no annotations, no
   `@Transactional` magic. Transaction boundaries become explicit
   `db.Transaction(func(tx *gorm.DB) error {...})` scopes (GORM, not raw
   `pgx.Tx` — see item 1); auth checks become explicit middleware.
3. **JVM/Tomcat/WAR → native binary.** No servlet container, no JSP, no warm-up.
   This is the operational payoff (startup, memory, image size) — quantified in
   [baseline-performance.md](baseline-performance.md).

**Deliberately unchanged:** PostgreSQL, the schema, Liquibase, the React
frontend, and the Playwright parity suite — so Java and Go can run side-by-side
against one database throughout the migration.

---

## 5. Java runtime specifics (JDK / JVM / Tomcat)

Measured from the running `openelisglobal-webapp` container (image
`itechuw/openelis-global-2:develop`).

### JDK

| Property       | Value                                                                                        |
| -------------- | -------------------------------------------------------------------------------------------- |
| Distribution   | **Eclipse Temurin (Adoptium)**                                                               |
| Version        | **OpenJDK 21.0.11+10 LTS**                                                                   |
| VM             | HotSpot **64-Bit Server VM**, mixed mode, **CDS/AppCDS sharing** enabled                     |
| Compile target | `--release 21` (`maven.java.release=21` in `pom.xml`)                                        |
| OS / arch      | Linux x86-64, container base has **12 vCPU** visible                                         |
| Timezone       | **`Etc/UTC`** (`/etc/timezone`) → this is the value `rest/server-time` returns as `timezone` |

### JVM memory & GC (ergonomic — no explicit `-Xmx`)

| Property               | Value                                        | Note                                                    |
| ---------------------- | -------------------------------------------- | ------------------------------------------------------- |
| Garbage collector      | **G1GC** (ergonomic default)                 | Go uses its own concurrent GC — no flags, no tuning.    |
| Initial heap           | ~376 MiB (`InitialHeapSize`)                 | Ergonomic (¼ of detected RAM).                          |
| Max heap               | ~6.0 GiB (`MaxHeapSize` / `SoftMaxHeapSize`) | No explicit `-Xmx`; sized from host RAM.                |
| Container memory limit | **none** (`HostConfig.Memory = 0`)           | JVM sizes off host, not a cgroup cap.                   |
| Measured RSS           | webapp ~1.5 GiB, FHIR ~2.2 GiB               | See [baseline-performance.md](baseline-performance.md). |

### Key JVM launch flags (Tomcat `CATALINA_OPTS` + app `-D`)

```
# module access (Hibernate/Spring reflection on JDK 21)
--add-opens=java.base/java.lang=ALL-UNNAMED
--add-opens=java.base/java.lang.reflect=ALL-UNNAMED
--add-opens=java.base/java.io=ALL-UNNAMED
--add-opens=java.base/java.util=ALL-UNNAMED
--add-opens=java.base/java.util.concurrent=ALL-UNNAMED
--add-opens=java.rmi/sun.rmi.transport=ALL-UNNAMED
# platform
-Djava.util.logging.manager=org.apache.juli.ClassLoaderLogManager
-Djdk.tls.ephemeralDHKeySize=2048
-Dsun.io.useCanonCaches=false
-Dorg.apache.catalina.security.SecurityListener.UMASK=0027
# app config (injected, not in code)
-Ddatasource.url=jdbc:postgresql://db.openelis.org:5432/clinlims
-Ddatasource.username=clinlims  -Ddatasource.password=***
-Doe.ssl.keystorepath=/etc/openelis-global/keystore   (kspass)
-Doe.ssl.truststorepath=/etc/openelis-global/truststore (tspass)
```

`--add-opens` exists purely because Hibernate/Spring do deep reflection under
JDK 21's module system — **none of this has a Go analog**; Go has no module
access flags, no logging-manager wiring, no reflection-open requirements.

### Servlet container / connectors

| Property      | Value                                                                                  |
| ------------- | -------------------------------------------------------------------------------------- |
| Server        | **Apache Tomcat 10.1.57**                                                              |
| Native        | Tomcat Native (APR) **2.0.15**, APR **1.7.2**, **OpenSSL 3.0.13**                      |
| TLS connector | `https-openssl-nio-8443`, RSA cert (keystore alias `tomcat`), **HTTP/2 via ALPN (h2)** |
| Config roots  | `catalina.base` / `catalina.home` = `/usr/local/tomcat`, tmp `/usr/local/tomcat/temp`  |

In the Go target, TLS is already terminated at the **nginx proxy** (ports
80/443), so the Go service speaks plain HTTP behind it — no keystore/truststore,
no OpenSSL/APR, no connector config.

### Migration implications

- **No `--add-opens` / module system.** Go's reflection needs no access grants;
  a whole class of JDK-21 launch config disappears.
- **No heap/GC tuning surface.** No `-Xmx`, no G1 vs ZGC choice; Go's GC is
  automatic. Memory footprint is expected to be far lower (baseline doc).
- **App config via `-D` → env/flags.** `datasource.*` and `oe.ssl.*` become
  environment variables / config in Go (`OE_GO_ADDR` is the first of these).
- **TLS/keystore handled by nginx, not the app.** The Go binary drops the entire
  Tomcat SSL/OpenSSL/APR stack.
- **`Etc/UTC`** is the container zone — the Go port must resolve the same IANA
  zone id (via `TZ`) for `server-time` parity, not Go's zone abbreviation.
