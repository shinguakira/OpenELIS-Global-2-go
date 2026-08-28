-- =============================================================================
-- SampleEdit form-load E2E fixture
-- =============================================================================
-- `SampleEditController.getSampleItems` is NOT a plain lookup: it calls
--
--     sampleItemService.getSampleItemsBySampleIdAndStatus(
--         sample.getId(), ENTERED_STATUS_SAMPLE_LIST)
--
-- and ENTERED_STATUS_SAMPLE_LIST holds exactly one id — the status_of_sample row
-- named `SampleEntered` (SampleStatus.Entered, resolved through IStatusService).
--
-- NOT ONE sample_item in the stock dev/demo dataset carries that status: the
-- whole table is status 1 and 28. So for every accession in the database the
-- filtered list comes back EMPTY, and three separate parts of the response are
-- pinned to their fallback values no matter what the server does:
--
--     existingTests      always []
--     possibleTests      always []
--     maxAccessionNumber always "<accession>-0"   (the else-branch; the real
--                                                  branch appends the LAST
--                                                  sample item's sort order)
--
-- A port that ignored the status filter entirely, or that returned every sample
-- item, would produce identical output on this dataset and pass. This fixture
-- seeds the one shape that tells those implementations apart: an accession whose
-- items ARE in SampleEntered status, plus a second accession whose items are
-- not, so the filter has something to exclude.
--
-- ---- WHAT EACH ROW IS FOR --------------------------------------------------
--
--   E2E-EDIT-01   TWO sample items, BOTH SampleEntered, sort orders 1 and 2,
--                 each carrying one non-canceled analysis.
--                 -> existingTests populated
--                 -> maxAccessionNumber = "E2E-EDIT-01-2" (the real branch)
--
--   E2E-EDIT-02   TWO sample items, one SampleEntered (sort 1) and one
--                 SampleDisposed (sort 2).
--                 -> maxAccessionNumber = "E2E-EDIT-02-1", NOT "-2": the
--                    excluded item is the LAST one, so a port that skips the
--                    status filter lands on "-2" and fails. This is the row
--                    that makes the filter observable rather than merely
--                    present.
--
-- The analyses are seeded in a status OTHER than `Test Canceled`, because
-- addCurrentTestsToList passes excludedAnalysisStatusList — {Canceled} — to
-- getAnalysesBySampleItemsExcludingByStatusIds. A canceled analysis would leave
-- existingTests empty again and defeat the point.
--
-- ---- ID / SEQUENCE POLICY --------------------------------------------------
-- nextval everywhere, cleaned up by accession marker. Reserved ids are not an
-- option for sample / sample_item / analysis: the loader's normalize_sequences
-- step runs setval(seq, MAX(id) + 1) over those tables.
--
-- lastupdated is set on every sample row. Hibernate optimistic-locks Sample on
-- that column, so a NULL there makes any dirty-check flush issue
-- `update SAMPLE ... where ID=? and LASTUPDATED=?`, match zero rows and throw
-- StaleStateException — which took rest/order/dashboard from 200 to 500 across
-- the whole table when an earlier fixture omitted it.
--
-- Usage (via the repo's loader, from repo root):
--   ./src/test/resources/load-test-fixtures.sh --profile=core
--
-- IDEMPOTENT: safe to re-run; every row is deleted by marker before re-insert.
-- =============================================================================

DO $BODY$
DECLARE
    target_patient  NUMERIC;
    tos_id          NUMERIC;
    entered_status  NUMERIC;
    other_status    NUMERIC;
    ana_status      NUMERIC;
    test_a          NUMERIC;
    test_b          NUMERIC;

    s_edit1         NUMERIC;
    s_edit2         NUMERIC;
    it_1a           NUMERIC;
    it_1b           NUMERIC;
    it_2a           NUMERIC;
    it_2b           NUMERIC;
BEGIN
    -- ---- cleanup, children first -------------------------------------------
    DELETE FROM clinlims.analysis
     WHERE sampitem_id IN (
        SELECT si.id FROM clinlims.sample_item si
          JOIN clinlims.sample s ON s.id = si.samp_id
         WHERE s.accession_number LIKE 'E2E-EDIT-%');
    DELETE FROM clinlims.sample_item
     WHERE samp_id IN (SELECT id FROM clinlims.sample
                        WHERE accession_number LIKE 'E2E-EDIT-%');
    DELETE FROM clinlims.sample_human
     WHERE samp_id IN (SELECT id FROM clinlims.sample
                        WHERE accession_number LIKE 'E2E-EDIT-%');
    DELETE FROM clinlims.sample WHERE accession_number LIKE 'E2E-EDIT-%';

    -- ---- reference data, resolved by NAME at load time ----------------------
    -- Resolved by name rather than by the ids this database happens to use
    -- (SampleEntered = 20 here), because IStatusService resolves them by name
    -- too and a differently-seeded deployment numbers them differently.
    SELECT id INTO entered_status FROM clinlims.status_of_sample
     WHERE status_type = 'SAMPLE' AND name = 'SampleEntered' LIMIT 1;
    SELECT id INTO other_status FROM clinlims.status_of_sample
     WHERE status_type = 'SAMPLE' AND name = 'SampleDisposed' LIMIT 1;
    SELECT id INTO ana_status FROM clinlims.status_of_sample
     WHERE status_type = 'ANALYSIS' AND name <> 'Test Canceled' ORDER BY id LIMIT 1;
    SELECT id INTO tos_id FROM clinlims.type_of_sample WHERE is_active = 'Y' ORDER BY id LIMIT 1;
    SELECT id INTO test_a FROM clinlims.test WHERE is_active = 'Y' ORDER BY id LIMIT 1;
    SELECT id INTO test_b FROM clinlims.test WHERE is_active = 'Y' ORDER BY id OFFSET 1 LIMIT 1;
    SELECT id INTO target_patient FROM clinlims.patient ORDER BY id LIMIT 1;

    IF entered_status IS NULL OR other_status IS NULL OR ana_status IS NULL
       OR tos_id IS NULL OR test_a IS NULL OR test_b IS NULL THEN
        RAISE NOTICE 'sample-edit-e2e: prerequisites missing; nothing seeded.';
        RETURN;
    END IF;

    -- ---- samples ------------------------------------------------------------
    s_edit1 := nextval('clinlims.sample_seq');
    s_edit2 := nextval('clinlims.sample_seq');
    INSERT INTO clinlims.sample
        (id, accession_number, entered_date, received_date, collection_date, lastupdated, is_confirmation)
    VALUES
        (s_edit1, 'E2E-EDIT-01', now(), now(), TIMESTAMP '2025-06-01 09:00:00', now(), false),
        (s_edit2, 'E2E-EDIT-02', now(), now(), TIMESTAMP '2025-06-02 09:00:00', now(), false);

    IF target_patient IS NOT NULL THEN
        INSERT INTO clinlims.sample_human (id, samp_id, patient_id) VALUES
            (nextval('clinlims.sample_human_seq'), s_edit1, target_patient),
            (nextval('clinlims.sample_human_seq'), s_edit2, target_patient);
    END IF;

    -- ---- sample items -------------------------------------------------------
    it_1a := nextval('clinlims.sample_item_seq');
    it_1b := nextval('clinlims.sample_item_seq');
    it_2a := nextval('clinlims.sample_item_seq');
    it_2b := nextval('clinlims.sample_item_seq');
    INSERT INTO clinlims.sample_item
        (id, samp_id, sort_order, typeosamp_id, status_id, collection_date, voided, rejected, lastupdated)
    VALUES
        -- E2E-EDIT-01: both SampleEntered -> maxAccessionNumber ends "-2"
        (it_1a, s_edit1, 1, tos_id, entered_status, now(), FALSE, FALSE, now()),
        (it_1b, s_edit1, 2, tos_id, entered_status, now(), FALSE, FALSE, now()),
        -- E2E-EDIT-02: the LAST item by sort order is NOT SampleEntered, so the
        -- filtered list ends at sort order 1 and maxAccessionNumber ends "-1".
        (it_2a, s_edit2, 1, tos_id, entered_status, now(), FALSE, FALSE, now()),
        (it_2b, s_edit2, 2, tos_id, other_status,   now(), FALSE, FALSE, now());

    -- ---- analyses -----------------------------------------------------------
    -- Non-canceled, so getAnalysesBySampleItemsExcludingByStatusIds keeps them
    -- and existingTests is actually populated.
    INSERT INTO clinlims.analysis
        (id, sampitem_id, test_id, status_id, analysis_type, entry_date, is_reportable, revision, lastupdated)
    VALUES
        (nextval('clinlims.analysis_seq'), it_1a, test_a, ana_status, 'MANUAL', now(), 'N', 0, now()),
        (nextval('clinlims.analysis_seq'), it_1b, test_b, ana_status, 'MANUAL', now(), 'N', 0, now()),
        (nextval('clinlims.analysis_seq'), it_2a, test_a, ana_status, 'MANUAL', now(), 'N', 0, now());

    RAISE NOTICE 'sample-edit-e2e: seeded E2E-EDIT-01 (2 entered items) and E2E-EDIT-02 (1 entered, 1 disposed)';
END $BODY$;
