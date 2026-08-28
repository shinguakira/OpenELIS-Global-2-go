-- =============================================================================
-- Order-search voided sample item — E2E fixture
-- =============================================================================
-- Fills a coverage gap proven by mutation testing, not guessed at.
--
-- `SampleItemServiceImpl.getSampleItemsBySampleId` builds the Hibernate
-- criteria {sample.id, voided:false}, so `GET rest/order/search` must exclude
-- voided sample items. Every sample_item row in the stock dev/demo dataset
-- (and in testdata/storage-e2e.xml) has voided = FALSE, so that filter is
-- UNOBSERVABLE: a port that drops the predicate entirely returns the identical
-- list and the c2 parity test passes anyway.
--
-- That was confirmed by deleting `AND si.voided = false` from the Go DAO and
-- re-running the c2 suite: it stayed green. This fixture is what makes the
-- mutant die.
--
-- Shape: one sample, THREE sample items — two live, one voided. A port that
-- forgets the filter returns 3 instead of 2 and fails
-- "order/search: voided sample items are excluded".
--
-- Why a dedicated sample instead of adding a voided item to E2E001: E2E001's
-- item list is asserted by the storage specs, so mutating it would couple two
-- unrelated suites. This row is invisible to them — the storage specs select
-- by their own accessions, and c1's anySampleAccession() INNER JOINs
-- sample_human on patients that already carry media.
--
-- Usage (via the repo's loader, from repo root):
--   ./src/test/resources/load-test-fixtures.sh --profile=core
--
-- The loader runs this in load_profile_lane_fixtures(), i.e. AFTER the storage
-- fixture, because the sample_human link attaches to the first patient by id
-- (patient 1000 from testdata/storage-e2e.xml).
--
-- IDEMPOTENT: keyed on the ACCESSION NUMBER and deleted before re-insert.
-- Deliberately NOT keyed on a reserved id: the loader's normalize_sequences
-- step runs setval('sample_seq', MAX(id) + 1), so a 9.9M id would permanently
-- drag the whole sample sequence up. Same reasoning as the E2E-NOPAT-01 row in
-- patient-media-e2e.sql.
-- =============================================================================

DO $$
DECLARE
    target_patient  NUMERIC;
    new_sample_id   NUMERIC;
    sample_type_id  NUMERIC;
    sample_status   NUMERIC;
    voided_item_id  NUMERIC;
    analysis_status NUMERIC;
    analysis_test   NUMERIC;
BEGIN
    -- Cleanup FIRST, before the patient is chosen: the choice below is a
    -- count over sample_human, and this fixture's own row from a previous run
    -- would otherwise be counted, letting the winner drift between loads.
    -- Children first — sample_item and sample_human FK-reference sample, and
    -- analysis FK-references sample_item.
    DELETE FROM clinlims.analysis
     WHERE sampitem_id IN (
        SELECT si.id FROM clinlims.sample_item si
          JOIN clinlims.sample s ON s.id = si.samp_id
         WHERE s.accession_number = 'E2E-VOIDED-01');
    DELETE FROM clinlims.sample_item
     WHERE samp_id IN (SELECT id FROM clinlims.sample
                        WHERE accession_number = 'E2E-VOIDED-01');
    DELETE FROM clinlims.sample_human
     WHERE samp_id IN (SELECT id FROM clinlims.sample
                        WHERE accession_number = 'E2E-VOIDED-01');
    DELETE FROM clinlims.sample WHERE accession_number = 'E2E-VOIDED-01';

    -- The patient with the MOST sample_human rows, which is exactly the patient
    -- c1-patient-reads.spec.ts's merge/details test selects. Not a coincidence
    -- to leave implicit: PatientMergeServiceImpl derives BOTH totalSamples and
    -- totalResults by walking getSampleItemsBySampleId, so the voided item has
    -- to land on that patient or those two filters stay unobservable there.
    --
    -- Ties break by id so the pick is deterministic, and since this fixture
    -- then adds one more row to the winner, that patient ends up a STRICT
    -- maximum — which also settles the tie the spec's own
    -- `ORDER BY count(*) DESC LIMIT 1` would otherwise resolve arbitrarily.
    SELECT sh.patient_id INTO target_patient
      FROM clinlims.sample_human sh
     GROUP BY sh.patient_id
     ORDER BY count(*) DESC, sh.patient_id
     LIMIT 1;

    -- Resolved at load time rather than hardcoded: these are seeded by
    -- ConfigurationInitializationService at app startup, so their ids differ
    -- between a fresh stack and a migrated one.
    SELECT id INTO sample_type_id FROM clinlims.type_of_sample ORDER BY id LIMIT 1;
    SELECT id INTO sample_status  FROM clinlims.status_of_sample ORDER BY id LIMIT 1;
    SELECT id INTO analysis_test  FROM clinlims.test WHERE is_active = 'Y' ORDER BY id LIMIT 1;
    -- MUST be an ANALYSIS status that countResultsForPatient does NOT exclude.
    -- The dev dataset's analyses are all "Not Tested", which IS excluded, so an
    -- excluded status here would leave totalResults at 0 either way and the
    -- voided filter would still be invisible on that field.
    SELECT id INTO analysis_status FROM clinlims.status_of_sample
     WHERE status_type = 'ANALYSIS'
       AND name NOT IN ('Test Canceled', 'Sample Rejected', 'Not Tested')
     ORDER BY id LIMIT 1;

    IF target_patient IS NULL OR sample_type_id IS NULL OR sample_status IS NULL THEN
        RAISE NOTICE 'order-search-e2e: prerequisites missing; nothing seeded.';
        RETURN;
    END IF;

    new_sample_id := nextval('clinlims.sample_seq');

    INSERT INTO clinlims.sample
        (id, accession_number, entered_date, received_date, lastupdated, is_confirmation)
    VALUES
        (new_sample_id, 'E2E-VOIDED-01', now(), now(), now(), false);

    INSERT INTO clinlims.sample_human (id, samp_id, patient_id)
    VALUES (nextval('clinlims.sample_human_seq'), new_sample_id, target_patient);

    -- Two live items and one voided one. The voided row sits in the MIDDLE by
    -- sort order so that a port which slices the list rather than filtering it
    -- also fails, instead of accidentally returning the right two.
    voided_item_id := nextval('clinlims.sample_item_seq');
    INSERT INTO clinlims.sample_item
        (id, samp_id, sort_order, typeosamp_id, status_id, collection_date,
         collector, quantity, voided, rejected, lastupdated)
    VALUES
        (nextval('clinlims.sample_item_seq'), new_sample_id, 1, sample_type_id,
         sample_status, now(), 'Tech-E2E', 5.0, FALSE, FALSE, now()),
        (voided_item_id, new_sample_id, 2, sample_type_id,
         sample_status, now(), 'Tech-E2E', 5.0, TRUE,  FALSE, now()),
        (nextval('clinlims.sample_item_seq'), new_sample_id, 3, sample_type_id,
         sample_status, now(), 'Tech-E2E', 5.0, FALSE, FALSE, now());

    -- One analysis hanging off the VOIDED item, in a status totalResults would
    -- otherwise count. This is what makes the voided filter observable on
    -- totalResults as well as on totalSamples: Java never walks the voided
    -- item, so it reports 0; a port that counts analyses straight from
    -- sample_item reports 1.
    IF analysis_test IS NOT NULL AND analysis_status IS NOT NULL THEN
        INSERT INTO clinlims.analysis
            (id, sampitem_id, test_id, status_id, analysis_type,
             entry_date, is_reportable, revision, lastupdated)
        VALUES
            (nextval('clinlims.analysis_seq'), voided_item_id, analysis_test,
             analysis_status, 'MANUAL', now(), 'N', 0, now());
    ELSE
        RAISE NOTICE 'order-search-e2e: no usable test/status; totalResults coverage skipped.';
    END IF;

    RAISE NOTICE 'order-search-e2e: seeded E2E-VOIDED-01 (sample %) on patient % with 2 live + 1 voided item',
        new_sample_id, target_patient;
END $$;
