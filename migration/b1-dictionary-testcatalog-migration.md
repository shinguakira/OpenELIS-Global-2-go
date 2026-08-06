# b1 — Dictionary + Test Catalog reference reads (scoped migration plan)

Status: **All stages complete ✅**
Branch: `migration/b1-dictionary-testcatalog` (forked from `migration-base`).
Companion docs:
[a2-static-reads-migration.md](a2-static-reads-migration.md),
[endpoint-migration-taxonomy.md](endpoint-migration-taxonomy.md),
[branch-naming.md](branch-naming.md).

b1 is "Wave 1 of Type B" — the highest-leverage reference reads.  Everything
downstream (samples, analyses, results) joins to this data, so correctness here
is a prerequisite for every subsequent wave.

---

## 0. Scope — 6 endpoints

| # | Endpoint | Java controller | Auth (Java) |
|---|----------|-----------------|-------------|
| 1 | `GET /rest/dictionary-categories` | `DictionaryMenuRestController.fetchDictionaryCategories()` | `@PreAuthorize("hasRole('ADMIN')")` class-level |
| 2 | `GET /rest/uom[?type=]` | `UnitOfMeasureRestController.getUnitOfMeasuresByType()` | **public** (no auth on GET) |
| 3 | `GET /rest/test-catalog/lab-units` | `TestCatalogEditorRestController.listLabUnits()` | `@PreAuthorize("hasRole('ADMIN')")` class-level |
| 4 | `GET /rest/test-catalog/sample-types` | `TestCatalogEditorRestController.listSampleTypes()` | `@PreAuthorize("hasRole('ADMIN')")` class-level |
| 5 | `GET /rest/test-catalog/panels` | `TestCatalogEditorRestController.listPanels()` | `@PreAuthorize("hasRole('ADMIN')")` class-level |
| 6 | `GET /rest/TestCatalog` | `TestCatalogRestController.showTestCatalog()` | `@PreAuthorize("hasRole('ADMIN')")` class-level |

Go serves all 6 without auth during the migration window (strangler-fig;
nginx controls routing — only proxy the Go path to authorized clients).

---

## 1. Per-endpoint Java source investigation

### 1. `dictionary-categories`

- Java: `DictionaryMenuRestController.fetchDictionaryCategories()`
  → `dictionaryCategoryService.getAll()` (BaseObjectService.getAll — no ORDER BY)
  → returns `List<DictionaryCategory>` serialized by Jackson.
- `DictionaryCategory` extends `BaseObject` which has `@JsonInclude(NON_NULL)` at
  the class level — `lastupdated` is omitted from the JSON when null.
- Table: `clinlims.dictionary_category`
  (`id`, `description`, `local_abbrev` → `localAbbreviation`, `name` → `categoryName`,
  `lastupdated` → epoch ms when non-null).
- No ORDER BY — DB-natural order; e2e does NOT assert ordering.

### 2. `uom[?type=]`

Two code paths in `UnitOfMeasureRestController.getUnitOfMeasuresByType()`:

| `?type=` | Java path | SQL |
|----------|-----------|-----|
| absent / blank | `unitOfMeasureService.getAll()` | `SELECT * FROM unit_of_measure` |
| present | `UomTypeMapDAOImpl.getUnitOfMeasuresByType(type)` | JOIN `uom_type_map` WHERE `uom_type = :type` |

Response shape: `List<IdValuePair>` = `[{"id":"...","value":"..."}]`
where `value` = `unitOfMeasureName`.

`uom_type_map` is seeded by Liquibase (`018-uom-type-mapping-table.xml`) with
exactly **4 UOMs for `SAMPLE_COLLECTION`**: mL, µL, tubes, slides.
Unknown type → empty list (no error).
No ORDER BY on either path — e2e does NOT assert ordering.

### 3. `test-catalog/lab-units`

- Java: `testSectionService.getAllTestSections()` (all rows, `test_section` table)
  → each `TestSection.getLocalizedName()` resolves English name via
  `localization_value JOIN` (locale = `'en'`), falls back to `name` column.
- **Sort:** `options.sort((a,b) -> a.name.compareToIgnoreCase(b.name))` in the
  controller — **case-insensitive ascending on the display name**.
- Response: `List<IdValuePair>` = `[{"id":"...","name":"..."}]`.
- e2e DOES assert case-insensitive ascending order.

### 4. `test-catalog/sample-types`

- Java: `typeOfSampleService.getAllTypeOfSamplesSortOrdered()`
  → rows ordered by `sort_order` column (ascending numeric, hidden from response).
- Display name: `getLocalizedName()` → `description` when non-blank, else
  `localAbbreviation` (no localization_value join — `type_of_sample.description`
  is the plain text name).
- Response: `List<IdValuePair>` = `[{"id":"...","name":"..."}]`.
- e2e does NOT assert ordering (sort_order is a hidden column not in the response).

### 5. `test-catalog/panels`

- Java: `panelService.getAllActivePanels()`
  → `WHERE is_active = 'Y' ORDER BY panelName` (SQL ORDER BY the name column).
- Response: `List<IdValuePair>` = `[{"id":"...","name":"..."}]`.
- e2e DOES assert case-insensitive ascending order on panel name.

### 6. `TestCatalog`

Most complex endpoint. Java: `TestCatalogRestController.showTestCatalog()`
→ `createTestList()` which N+1-loops over every test calling:

```
testService.getAllTests(false)           → all test rows
testService.getTestSectionName(test)    → test_section.name / localized name
testService.getTypeOfSample(test)       → typeOfSampleTests.get(0) (first sample type)
testService.getResultType(test)         → test_result.result_type
testService.getUOM(test, false)         → unit_of_measure.name
testService.getPossibleTestResults(test)→ result type / significant digits
```

Then sorts by `(testUnit, sampleType, panel, testSortOrder)` (all String.compareTo,
case-sensitive ascending). `testSectionList` = distinct `testUnit` values extracted
in the sorted order.

Response shape: `TestCatalogForm`:
```json
{
  "formName": "testCatalogForm",
  "testSectionList": ["Chemistry", "Hematology", ...],
  "testCatalogList": [
    {
      "id": "1",
      "localization": {
        "id": "42",
        "description": "...",
        "values": {"en": {"id": "99", "locale": "en", "value": "Glucose"}},
        "lastupdated": 1234567890000
      },
      "testUnit": "Chemistry",
      "sampleType": "Blood",
      "panel": "None",
      "active": "Active",
      "orderable": "Orderable",
      "uom": "mg/dL",
      "significantDigits": "n/a",
      "hasLimitValues": false,
      "hasDictionaryValues": false
    }
  ]
}
```

`localization` is the Jackson serialization of `Localization.java` with
`@JsonInclude(NON_NULL)` — `description` and `lastupdated` omitted when null.
`values` is a `Map<String, LocalizationValue>` keyed by locale string.

---

## 2. DB dependencies

| Domain | Table(s) | Key columns |
|--------|----------|-------------|
| `dictionarycategory` | `dictionary_category` | `id`, `description`, `local_abbrev`, `name`, `lastupdated` |
| `unitofmeasure` | `unit_of_measure`, `uom_type_map` | `unit_of_measure.id/name`; `uom_type_map.uom_id/uom_type` |
| `test` (TestSection) | `test_section`, `localization_value` | `test_section.name_localization_id` → `localization_value.localization_id` WHERE `locale='en'` |
| `typeofsample` | `type_of_sample` | `id`, `description`, `local_abbrev`, `sort_order` |
| `panel` | `panel` | `id`, `name`, `is_active` |
| `test` (TestCatalog) | `test`, `test_section`, `sampletype_test`, `type_of_sample`, `unit_of_measure`, `localization`, `localization_value` | see §1.6; `sampletype_test` table name (NOT `type_of_sample_test`); column `sample_type_id` (NOT `type_of_samp_id`) |

**Critical schema notes:**
- The join table between test and sample type is **`sampletype_test`** (HBM
  `table="SAMPLETYPE_TEST"`), columns `test_id` and `sample_type_id`.
- A test can have multiple sample types; Java uses `.get(0)` — Go uses
  `LATERAL ... LIMIT 1` to replicate this.
- `type_of_sample` DOES have `name_localization_id` (from HBM) — the localization
  join is valid for sample type name resolution.

---

## 3. Go layer structure

Each domain follows the same 4-file pattern mirroring the Java source.
**SQL lives only in daoimpl; controller/rest contains only HTTP plumbing and DTO
conversion (no `database/sql` imports).**

```
internal/
  dictionarycategory/
    valueholder/dictionary_category.go         ← DictionaryCategory.java
    daoimpl/dictionary_category_dao_impl.go    ← DictionaryCategoryDAOImpl.java
    service/dictionary_category_service_impl.go← DictionaryCategoryServiceImpl.java
    controller/rest/dictionary_category.go     ← DictionaryMenuRestController.java (thin)

  unitofmeasure/
    valueholder/unit_of_measure.go             ← UnitOfMeasure.java
    valueholder/uom_type_map.go                ← UomTypeMap.java
    daoimpl/unit_of_measure_dao_impl.go        ← UnitOfMeasureDAOImpl.java
    daoimpl/uom_type_map_dao_impl.go           ← UomTypeMapDAOImpl.java
    service/unit_of_measure_service_impl.go    ← UnitOfMeasureServiceImpl.java
    controller/rest/uom.go                     ← UnitOfMeasureRestController.java (thin)

  test/
    valueholder/test_section.go                ← TestSection.java
    valueholder/test_catalog.go                ← TestCatalog.java
    daoimpl/test_section_dao_impl.go           ← TestSectionDAOImpl.java
    service/test_section_service_impl.go       ← TestSectionServiceImpl.java

  typeofsample/
    valueholder/type_of_sample.go              ← TypeOfSample.java
    daoimpl/type_of_sample_dao_impl.go         ← TypeOfSampleDAOImpl.java
    service/type_of_sample_service_impl.go     ← TypeOfSampleServiceImpl.java

  panel/
    valueholder/panel.go                       ← Panel.java
    daoimpl/panel_dao_impl.go                  ← PanelDAOImpl.java
    service/panel_service_impl.go              ← PanelServiceImpl.java

  localization/
    valueholder/localization.go                ← Localization.java + LocalizationValue.java
    (valueholder/supported_locale.go — a2)

  testcatalog/
    controller/rest/testcatalog_editor.go      ← TestCatalogEditorRestController.java (thin)
      calls: TestSectionService, TypeOfSampleService, PanelService

  testconfiguration/
    form/test_catalog_form.go                  ← TestCatalogForm.java
    service/test_catalog_service_impl.go       ← createTestList() logic extracted from controller
    controller/rest/test_catalog.go            ← TestCatalogRestController.java (thin)
```

`main.go` wires each domain via explicit DI:
```
DAO struct{DB} → service struct{DAO} → controller struct{Service} → Routes(mux, ctrl)
```

---

## 4. e2e parity gate

Spec: `migration/openelis-api-e2e/tests/readonly/b1-testcatalog.spec.ts`
Project: `api-readonly` (admin auth via `storageState`).

### Coverage gap found and fixed

`GET /rest/uom?type=` was not tested. Two tests added:
- `uom?type=SAMPLE_COLLECTION` — verifies `{id,value}` shape + count matches
  DB oracle (`uom_type_map` seeds exactly 4 rows for this type).
- `uom?type=NONEXISTENT` — verifies unknown type returns empty array with 200.

### Test matrix

| Test | What it asserts |
|------|-----------------|
| `dictionary-categories` | `{id,description,localAbbreviation,categoryName}` required; `lastupdated` optional number; unique ids |
| `uom` (unfiltered) | `{id,value}` exact keys; unique ids |
| `uom?type=SAMPLE_COLLECTION` | `{id,value}` shape; unique ids; count == DB oracle (4) |
| `uom?type=NONEXISTENT` | 200 + JSON array + length 0 |
| `test-catalog/lab-units` | `{id,name}` exact; unique ids; case-insensitive ascending name order |
| `test-catalog/sample-types` | `{id,name}` exact; unique ids; order NOT asserted |
| `test-catalog/panels` | `{id,name}` exact; unique ids; case-insensitive ascending name order |
| `TestCatalog` | `formName=="testCatalogForm"`; `testSectionList` non-empty, unique, ascending; `testCatalogList` non-empty; each row has `id` + `localization` object |
| DB oracle | `test` count, `dictionary` count, `type_of_sample` count, `uom_type_map SAMPLE_COLLECTION` count == 4 |

`go-parity` testMatch does NOT include b1 (5 of 6 endpoints require admin auth
in Java; go-parity has no auth). Add b1 to go-parity when auth middleware is
ported.

---

## 5. Commit history

| # | Commit | Description |
|---|--------|-------------|
| 1 | `53aa934` | Initial b1 Go port (monolithic — all SQL in controller; **superseded by commit 2**) |
| 2 | `f8ac5f1` | Refactor: apply full layered architecture (valueholder/daoimpl/service/controller); fix `sampletype_test` table name and `sample_type_id` column name; add LATERAL LIMIT 1 for sample type |
