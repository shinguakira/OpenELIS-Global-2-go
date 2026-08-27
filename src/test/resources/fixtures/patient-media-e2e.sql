-- =============================================================================
-- Patient media (photos + ID documents) — E2E fixture
-- =============================================================================
-- Fills a real coverage gap: clinlims.patient_photo, clinlims.patient_id_document
-- and clinlims.image are ALL EMPTY in the stock dev/demo dataset, so the
-- c1-patient-reads parity tests can only exercise the "no media" path. Four
-- behaviors of the Java API are therefore untested without this file:
--
--   1. GET /rest/patient-photos/{id}/false  returns a full data-URI
--      ("data:image/png;base64,...") while
--      GET /rest/patient-photos/{id}/true   returns BARE base64 with no prefix.
--      (PatientPhotoServiceImpl.java:116-119 — the single biggest parity trap
--      in this endpoint group; the frontend compensates for it, so a port that
--      emits a data-URI for both silently breaks the patient avatar.)
--   2. GET /rest/patient-id-documents/{id} returns a populated list, whose
--      `description` and `lastUpdated` keys are OMITTED (not null) when the
--      column is null — Jackson NON_NULL content inclusion.
--   3. GET /rest/patient-id-documents/{id}/{docId}/full returns that document's
--      own data-URI, and returns {"data":""} for a docId belonging to a
--      different patient (no 403/404 — see the IDOR-shaped note in
--      migration/c1-patient-reads-migration notes).
--   4. The DAO's `deleted = false` filter actually hides soft-deleted rows.
--
-- Usage:
--   psql -U clinlims -d clinlims -f patient-media-e2e.sql
-- or, via the repo's loader (from repo root):
--   ./src/test/resources/load-test-fixtures.sh --profile=core
--
-- The loader runs this automatically, in load_profile_lane_fixtures() — i.e.
-- AFTER the storage fixture, because the rows below attach to the first patient
-- by id, which storage-e2e.xml provides (patient 1000). It is loaded for both
-- profiles and is fatal on error: when it does not run, four c1 parity tests
-- silently take their test.skip branch instead of failing.
--
-- =============================================================================
-- IDEMPOTENT: safe to re-run. Rows are keyed on fixed ids in a reserved range
-- (9900000+) and deleted before re-insert, so this never duplicates and never
-- touches real seeded data.
-- =============================================================================

-- Attach to the lowest-numbered existing patient, which is the same patient the
-- c1 parity spec selects via `ORDER BY id LIMIT 1`. Resolved dynamically rather
-- than hardcoded so this works against any dataset.
DO $$
DECLARE
    target_patient  VARCHAR;
    other_patient   VARCHAR;
    -- A real, decodable 1x1 transparent PNG. Kept tiny on purpose: these are
    -- TEXT columns and the payload is echoed back in full by the API, so a
    -- large blob would bloat every response the parity suite diffs.
    png_b64         TEXT := 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==';
    -- Deliberately DIFFERENT from photo_data so a test can prove the thumbnail
    -- branch really reads thumbnail_data and not photo_data.
    thumb_b64       TEXT := 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADElEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==';
BEGIN
    SELECT id INTO target_patient FROM clinlims.patient ORDER BY id LIMIT 1;
    SELECT id INTO other_patient  FROM clinlims.patient ORDER BY id OFFSET 1 LIMIT 1;

    IF target_patient IS NULL THEN
        RAISE NOTICE 'patient-media-e2e: no patients in this database; nothing seeded.';
        RETURN;
    END IF;

    -- ---------------------------------------------------------------------
    -- Clean prior runs — BOUNDED to the reserved block, not `>= 9900000`.
    --
    -- An open-ended cleanup was actively destructive here, because the same
    -- block below advances both sequences to 9900100: after one fixture run
    -- every ordinary application insert receives an id of 9900101 or higher,
    -- which `>= 9900000` matches. Re-running test setup would then delete
    -- real patient photos and ID documents created since the last run.
    -- The reserved block is 9900000-9900099 and the sequences start above it,
    -- so the two ranges cannot overlap.
    -- ---------------------------------------------------------------------
    DELETE FROM clinlims.patient_photo        WHERE id BETWEEN 9900000 AND 9900099;
    DELETE FROM clinlims.patient_id_document  WHERE id BETWEEN 9900000 AND 9900099;

    -- ---------------------------------------------------------------------
    -- PHOTO — exactly one row (patient_id is UNIQUE on this table).
    -- photo_data and thumbnail_data differ so the two API branches are
    -- distinguishable from each other, not just from empty.
    -- ---------------------------------------------------------------------
    INSERT INTO clinlims.patient_photo (id, patient_id, photo_data, thumbnail_data, photo_type, last_updated)
    VALUES (9900001, target_patient, png_b64, thumb_b64, 'image/png', now());

    -- ---------------------------------------------------------------------
    -- ID DOCUMENTS — three rows covering three distinct assertions:
    --   9900001  description SET      -> `description` key PRESENT
    --   9900002  description NULL     -> `description` key ABSENT (NON_NULL)
    --   9900003  deleted = true       -> must NOT appear in the list at all
    -- ---------------------------------------------------------------------
    INSERT INTO clinlims.patient_id_document
        (id, patient_id, document_data, thumbnail_data, document_type, document_category, description, deleted, last_updated)
    VALUES
        (9900001, target_patient, png_b64, thumb_b64, 'image/png', 'ID_CARD', 'E2E national ID card', false, now()),
        (9900002, target_patient, png_b64, thumb_b64, 'image/png', 'PASSPORT', NULL,                   false, now()),
        (9900003, target_patient, png_b64, thumb_b64, 'image/png', 'ID_CARD', 'E2E soft-deleted doc',  true,  now());

    -- A document owned by a DIFFERENT patient, so a test can prove that
    -- requesting it under the wrong patientId yields {"data":""} rather than
    -- another patient's document.
    IF other_patient IS NOT NULL THEN
        INSERT INTO clinlims.patient_id_document
            (id, patient_id, document_data, thumbnail_data, document_type, document_category, description, deleted, last_updated)
        VALUES
            (9900004, other_patient, png_b64, thumb_b64, 'image/png', 'ID_CARD', 'E2E other-patient doc', false, now());
    END IF;

    -- Keep the sequences ahead of the reserved range so ordinary application
    -- inserts cannot collide with these fixed ids.
    PERFORM setval('clinlims.patient_photo_seq',
                   GREATEST(9900100, (SELECT last_value FROM clinlims.patient_photo_seq)), true);
    PERFORM setval('clinlims.patient_id_document_seq',
                   GREATEST(9900100, (SELECT last_value FROM clinlims.patient_id_document_seq)), true);

    RAISE NOTICE 'patient-media-e2e: seeded photo + documents for patient %', target_patient;
END $$;

-- =============================================================================
-- A sample with NO patient — the second, distinct 404 on patientByLabNumer
-- =============================================================================
-- SampleEditRestController.getPatientByLabNumber has TWO independent 404 paths:
--
--   Sample sample = getSample(accessionNumber);
--   if (sample == null)  return notFound();            <- unknown accession
--   Patient patient = sampleHumanService.getPatientForSample(sample);
--   if (patient == null) return notFound();            <- THIS one
--
-- The stock dataset has no sample without a sample_human row, so the second
-- branch is unreachable and a port implementing only the first passes every
-- test. This seeds one: a real sample row, deliberately with no sample_human,
-- so the accession RESOLVES but the patient lookup comes back empty.
--
-- Safe to add: no spec asserts a clinlims.sample row count, fixtures/db.ts's
-- BASELINE does not track samples, and c1-patient-reads.spec.ts's
-- anySampleAccession() helper INNER JOINs sample_human, so it can never pick
-- this row for the happy-path tests.
-- Keyed on the ACCESSION NUMBER, not a reserved id — unlike the media rows
-- above, and deliberately so.
--
-- `sample` is a core clinical table, and the loader's normalize_sequences step
-- runs `setval('sample_seq', MAX(id) + 1)` over it. A reserved id of 9900001
-- therefore does not stay contained: it drags the whole sample sequence from
-- ~1k to ~9.9M on every fixture load, permanently, for one throwaway row.
--
-- Letting the sequence assign the id and using the unique accession as the
-- cleanup key avoids that entirely, and is the create-then-cleanup-by-marker
-- pattern openelis-api-e2e.md §15 already prescribes for mutating fixtures.
DO $$
BEGIN
    DELETE FROM clinlims.sample WHERE accession_number = 'E2E-NOPAT-01';

    INSERT INTO clinlims.sample
        (id, accession_number, entered_date, received_date, is_confirmation)
    VALUES
        (nextval('clinlims.sample_seq'), 'E2E-NOPAT-01', now(), now(), false);

    RAISE NOTICE 'patient-media-e2e: seeded patient-less sample E2E-NOPAT-01';
END $$;
