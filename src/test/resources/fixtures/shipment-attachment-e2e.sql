-- =============================================================================
-- Unassigned-referral + order-attachment E2E fixture
-- =============================================================================
-- `clinlims.referral` and `clinlims.order_attachment` are BOTH EMPTY in the
-- stock dev/demo dataset. That left six c2 endpoints verifiable only down to
-- their envelope:
--
--     GET rest/unassigned-sample                     (dashboard rows)
--     GET rest/unassigned-sample/items               (SampleItemDTO rows)
--     GET rest/unassigned-sample/items/search        (same, filtered)
--     GET rest/unassigned-sample/by-facility/{id}
--     GET rest/order/{accessionNumber}/attachments   (the 200 path)
--     GET rest/order/attachments/{id}/download|view  (entirely)
--
-- `expect(Array.isArray(body)).toBe(true)` passes on `[]` forever, so the row
-- shape -- the only part that carries parity risk -- was never compared against
-- Java at all. This file supplies rows chosen so that each BRANCH in the Java
-- code produces a visibly different row, rather than merely making the arrays
-- non-empty.
--
-- ---- WHAT EACH ROW IS FOR --------------------------------------------------
--
-- Included (must appear):
--   R1  full row: organization, reason, priority, request date all set.
--   R4  a second referral on the SAME sample item as R1, with a LATER request
--       date, so referralTests[] has more than one entry and its ordering
--       (ORDER BY r.requestDate) is real rather than accidental.
--   R2  organization NULL but organization_name set -- exercises
--       compileSampleData's ELSE branch, which emits destinationFacilityName
--       WITHOUT destinationFacilityId. priority is NULL too, so the literal
--       "Normal" fallback shows up.
--   R5  on a different accession that has a SINGLE sample item.
--   R3  request date NULL and reason NULL -- daysUnassigned falls back to 0 and
--       BOTH referralDate and referralReasonId vanish from the row (the row is
--       a HashMap, so Jackson's NON_NULL drops null values).
--
-- Excluded (must NOT appear) -- one row per exclusion rule, because a port that
-- implements four of the five filters still returns a plausible-looking list:
--   X1  lost_status = true
--   X2  status = 'CANCELED'
--   X3  assigned to a shipping box (assigned_to_box_id set)
--   X4  status IS NULL. Subtle and easy to get wrong: the HQL says
--       `r.status != 'CANCELED'`, and in three-valued logic NULL != 'X' is
--       UNKNOWN, not TRUE -- so a NULL-status referral is excluded everywhere.
--       A port written with Go's `status != "CANCELED"` INCLUDES it.
--   X5  a VOIDED sample item. Deliberately asymmetric: the /items queries
--       filter `si.voided`, but the dashboard query does NOT, so this referral
--       appears in `unassigned-sample` and is absent from
--       `unassigned-sample/items`. A port that shares one filter between the
--       two endpoints gets exactly one of them wrong.
--
-- The two included accessions differ on purpose:
--   E2E-REF-01 has TWO sample items, so buildSampleItemDTOs suffixes the
--              accession with the sort order (E2E-REF-01-1, E2E-REF-01-2).
--   E2E-REF-02 has ONE, so it is NOT suffixed.
--   That rule is `accessionCounts > 1`, and on a single-item dataset it is
--   invisible.
--
-- Attachments on E2E-ATT-01:
--   A1  active, file_type set   -> listed; download/view use that type.
--   A2  active, file_type NULL  -> toDto emits "" while serveAttachment falls
--                                  back to application/octet-stream: the same
--                                  column, two different null policies.
--   A3  is_deleted = true       -> absent from the list AND 404 on
--                                  download/view.
--
-- ---- ID / SEQUENCE POLICY --------------------------------------------------
-- Everything uses nextval and is cleaned up by MARKER (accession number, box
-- id), never by a reserved id range. For sample / sample_item / sample_human /
-- analysis that is mandatory: the loader's normalize_sequences step runs
-- setval(seq, MAX(id)+1) over those tables, so a 9.9M id would permanently drag
-- the sequence up. referral, order_attachment, shipping_box and box_sample_item
-- are not normalized, but follow the same rule for consistency.
--
-- Usage (via the repo's loader, from repo root):
--   ./src/test/resources/load-test-fixtures.sh --profile=core
--
-- IDEMPOTENT: safe to re-run; every row is deleted by marker before re-insert.
-- =============================================================================

DO $BODY$
DECLARE
    order_status    NUMERIC;   -- status_of_sample, status_type='ORDER'
    org_a           NUMERIC;
    org_b           NUMERIC;
    ref_type        NUMERIC;
    ref_reason      NUMERIC;
    tos_id          NUMERIC;
    samp_status     NUMERIC;
    ana_status      NUMERIC;
    test_a          NUMERIC;
    test_b          NUMERIC;
    target_patient  NUMERIC;

    s_ref1          NUMERIC;
    s_ref2          NUMERIC;
    s_excl          NUMERIC;
    s_att           NUMERIC;

    it_1a           NUMERIC;
    it_1b           NUMERIC;
    it_2a           NUMERIC;
    it_lost         NUMERIC;
    it_cancel       NUMERIC;
    it_boxed        NUMERIC;
    it_nullstat     NUMERIC;
    it_voided       NUMERIC;

    an_1a           NUMERIC;
    an_1b           NUMERIC;
    an_1b_item      NUMERIC;
    an_2a           NUMERIC;
    an_lost         NUMERIC;
    an_cancel       NUMERIC;
    an_boxed        NUMERIC;
    an_nullstat     NUMERIC;
    an_voided       NUMERIC;

    box_pk          INTEGER;
    pdf_bytes       BYTEA;
    bin_bytes       BYTEA;
BEGIN
    -- sample.status_id is the ORDER-level status (status_type='ORDER'), NOT
    -- the SAMPLE-level one used by sample_item. Every stock sample carries it,
    -- and Java dereferences it without a null check, so leaving it NULL breaks
    -- unrelated endpoints (WorkPlanByTest 500s on the resulting NPE).
    SELECT id INTO order_status FROM clinlims.status_of_sample
     WHERE status_type = 'ORDER' AND name = 'Test Entered' LIMIT 1;
    -- ---- cleanup, children first -------------------------------------------
    DELETE FROM clinlims.box_sample_item
     WHERE sample_item_id IN (
        SELECT si.id::INTEGER FROM clinlims.sample_item si
          JOIN clinlims.sample s ON s.id = si.samp_id
         WHERE s.accession_number LIKE 'E2E-REF-%' OR s.accession_number = 'E2E-ATT-01')
        OR shipping_box_id IN (SELECT id FROM clinlims.shipping_box WHERE box_id = 'E2E-BOX-01');
    DELETE FROM clinlims.referral
     WHERE analysis_id IN (
        SELECT a.id FROM clinlims.analysis a
          JOIN clinlims.sample_item si ON si.id = a.sampitem_id
          JOIN clinlims.sample s ON s.id = si.samp_id
         WHERE s.accession_number LIKE 'E2E-REF-%' OR s.accession_number = 'E2E-ATT-01');
    DELETE FROM clinlims.order_attachment
     WHERE sample_id IN (SELECT id FROM clinlims.sample
                          WHERE accession_number LIKE 'E2E-REF-%' OR accession_number = 'E2E-ATT-01');
    DELETE FROM clinlims.analysis
     WHERE sampitem_id IN (
        SELECT si.id FROM clinlims.sample_item si
          JOIN clinlims.sample s ON s.id = si.samp_id
         WHERE s.accession_number LIKE 'E2E-REF-%' OR s.accession_number = 'E2E-ATT-01');
    DELETE FROM clinlims.sample_item
     WHERE samp_id IN (SELECT id FROM clinlims.sample
                        WHERE accession_number LIKE 'E2E-REF-%' OR accession_number = 'E2E-ATT-01');
    DELETE FROM clinlims.sample_human
     WHERE samp_id IN (SELECT id FROM clinlims.sample
                        WHERE accession_number LIKE 'E2E-REF-%' OR accession_number = 'E2E-ATT-01');
    DELETE FROM clinlims.shipping_box WHERE box_id = 'E2E-BOX-01';
    DELETE FROM clinlims.sample
     WHERE accession_number LIKE 'E2E-REF-%' OR accession_number = 'E2E-ATT-01';

    -- ---- reference data, resolved at load time ------------------------------
    -- Two DISTINCT organizations: by-facility/{id} is only a real filter if at
    -- least one included row belongs to a different facility.
    -- 送り先の組織2件。'ORDER BY id LIMIT/OFFSET' で素朴に選んでは いけない:
    -- organization は他の fixture も書き込むテーブルで、1件挿入されただけで
    -- どの組織が選ばれるかがずれる。実際 order-search-full-e2e.sql が
    -- E2E-DEPT を低い id で作った時に org_b がそれに化け、その fixture が次の
    -- ロードで delete+create した時点で referral.organization_id が NULL に落ち、
    -- 可視 referral が 10→9、施設が 2→1 になった。
    --
    -- このスイートが所有する組織（short_name が 'E2E-' 始まり）を除外して、
    -- デプロイ本来の組織だけから選ぶ。
    SELECT id INTO org_a FROM clinlims.organization
     WHERE short_name IS NULL OR short_name NOT LIKE 'E2E-%'
     ORDER BY id LIMIT 1;
    SELECT id INTO org_b FROM clinlims.organization
     WHERE short_name IS NULL OR short_name NOT LIKE 'E2E-%'
     ORDER BY id OFFSET 1 LIMIT 1;
    SELECT id INTO ref_type    FROM clinlims.referral_type   ORDER BY id LIMIT 1;
    SELECT id INTO ref_reason  FROM clinlims.referral_reason ORDER BY id LIMIT 1;
    SELECT id INTO tos_id      FROM clinlims.type_of_sample  ORDER BY id LIMIT 1;
    SELECT id INTO samp_status FROM clinlims.status_of_sample ORDER BY id LIMIT 1;
    SELECT id INTO ana_status  FROM clinlims.status_of_sample
                               WHERE status_type = 'ANALYSIS' ORDER BY id LIMIT 1;
    -- Two DISTINCT tests so referralTests[] entries can be told apart by name.
    SELECT id INTO test_a FROM clinlims.test WHERE is_active = 'Y' ORDER BY id LIMIT 1;
    SELECT id INTO test_b FROM clinlims.test WHERE is_active = 'Y' ORDER BY id OFFSET 1 LIMIT 1;
    SELECT id INTO target_patient FROM clinlims.patient ORDER BY id LIMIT 1;

    IF org_a IS NULL OR org_b IS NULL OR ref_type IS NULL OR tos_id IS NULL
       OR test_a IS NULL OR test_b IS NULL OR samp_status IS NULL THEN
        RAISE NOTICE 'shipment-attachment-e2e: prerequisites missing; nothing seeded.';
        RETURN;
    END IF;

    -- ---- samples ------------------------------------------------------------
    -- collection_date is set because SampleItemDTO.collectionDate reads it off
    -- the SAMPLE, not off the sample item.
    --
    -- lastupdated MUST be non-null. Hibernate optimistic-locks Sample on that
    -- column, so a row with lastupdated NULL makes any dirty-check flush issue
    -- `update SAMPLE ... where ID=? and LASTUPDATED=?`, match zero rows and
    -- throw StaleStateException. Leaving it null took rest/order/dashboard from
    -- 200 to 500 for the whole table, not just for these rows.
    s_ref1 := nextval('clinlims.sample_seq');
    s_ref2 := nextval('clinlims.sample_seq');
    s_excl := nextval('clinlims.sample_seq');
    s_att  := nextval('clinlims.sample_seq');
    INSERT INTO clinlims.sample
        (id, accession_number, entered_date, received_date, collection_date, lastupdated, is_confirmation, status_id)
    VALUES
        (s_ref1, 'E2E-REF-01', now(), now(), TIMESTAMP '2025-05-01 09:15:00', now(), false, order_status),
        (s_ref2, 'E2E-REF-02', now(), now(), TIMESTAMP '2025-05-02 09:15:00', now(), false, order_status),
        (s_excl, 'E2E-REF-03', now(), now(), TIMESTAMP '2025-05-03 09:15:00', now(), false, order_status),
        (s_att,  'E2E-ATT-01', now(), now(), TIMESTAMP '2025-05-04 09:15:00', now(), false, order_status);

    -- s_excl is deliberately left patient-less: the exclusion rows must not
    -- shift any patient-keyed count (patient/merge/details in particular --
    -- see order-search-e2e.sql for why that matters).
    IF target_patient IS NOT NULL THEN
        INSERT INTO clinlims.sample_human (id, samp_id, patient_id) VALUES
            (nextval('clinlims.sample_human_seq'), s_ref1, target_patient),
            (nextval('clinlims.sample_human_seq'), s_ref2, target_patient),
            (nextval('clinlims.sample_human_seq'), s_att,  target_patient);
    END IF;

    -- ---- sample items -------------------------------------------------------
    it_1a       := nextval('clinlims.sample_item_seq');
    it_1b       := nextval('clinlims.sample_item_seq');
    it_2a       := nextval('clinlims.sample_item_seq');
    it_lost     := nextval('clinlims.sample_item_seq');
    it_cancel   := nextval('clinlims.sample_item_seq');
    it_boxed    := nextval('clinlims.sample_item_seq');
    it_nullstat := nextval('clinlims.sample_item_seq');
    it_voided   := nextval('clinlims.sample_item_seq');
    INSERT INTO clinlims.sample_item
        (id, samp_id, sort_order, typeosamp_id, status_id, collection_date, voided, rejected, lastupdated)
    VALUES
        -- E2E-REF-01: TWO items -> accession gets suffixed with the sort order.
        (it_1a, s_ref1, 1, tos_id, samp_status, now(), FALSE, FALSE, now()),
        -- typeosamp_id NULL: the HQL LEFT JOINs type_of_sample and COALESCEs
        -- the description to '', while the id stays null -- so `typeOfSample`
        -- is "" and `typeOfSampleId` is absent. One column, two null policies.
        (it_1b, s_ref1, 2, NULL,   samp_status, now(), FALSE, FALSE, now()),
        -- E2E-REF-02: ONE item -> NOT suffixed.
        (it_2a, s_ref2, 1, tos_id, samp_status, now(), FALSE, FALSE, now()),
        -- exclusion carriers
        (it_lost,     s_excl, 1, tos_id, samp_status, now(), FALSE, FALSE, now()),
        (it_cancel,   s_excl, 2, tos_id, samp_status, now(), FALSE, FALSE, now()),
        (it_boxed,    s_excl, 3, tos_id, samp_status, now(), FALSE, FALSE, now()),
        (it_nullstat, s_excl, 4, tos_id, samp_status, now(), FALSE, FALSE, now()),
        (it_voided,   s_excl, 5, tos_id, samp_status, now(), TRUE,  FALSE, now());

    -- ---- analyses (one per referral) ----------------------------------------
    an_1a       := nextval('clinlims.analysis_seq');
    an_1b       := nextval('clinlims.analysis_seq');
    an_1b_item  := nextval('clinlims.analysis_seq');
    an_2a       := nextval('clinlims.analysis_seq');
    an_lost     := nextval('clinlims.analysis_seq');
    an_cancel   := nextval('clinlims.analysis_seq');
    an_boxed    := nextval('clinlims.analysis_seq');
    an_nullstat := nextval('clinlims.analysis_seq');
    an_voided   := nextval('clinlims.analysis_seq');
    INSERT INTO clinlims.analysis
        (id, sampitem_id, test_id, status_id, analysis_type, entry_date, is_reportable, revision, lastupdated)
    VALUES
        -- an_1a and an_1b BOTH hang off it_1a, which is what gives that one
        -- sample item two referralTests entries.
        (an_1a,       it_1a,       test_a, ana_status, 'MANUAL', now(), 'N', 0, now()),
        (an_1b,       it_1a,       test_b, ana_status, 'MANUAL', now(), 'N', 0, now()),
        (an_1b_item,  it_1b,       test_a, ana_status, 'MANUAL', now(), 'N', 0, now()),
        (an_2a,       it_2a,       test_a, ana_status, 'MANUAL', now(), 'N', 0, now()),
        (an_lost,     it_lost,     test_a, ana_status, 'MANUAL', now(), 'N', 0, now()),
        (an_cancel,   it_cancel,   test_a, ana_status, 'MANUAL', now(), 'N', 0, now()),
        (an_boxed,    it_boxed,    test_a, ana_status, 'MANUAL', now(), 'N', 0, now()),
        (an_nullstat, it_nullstat, test_a, ana_status, 'MANUAL', now(), 'N', 0, now()),
        (an_voided,   it_voided,   test_a, ana_status, 'MANUAL', now(), 'N', 0, now());

    -- ---- shipping box, for the "already assigned" exclusion ------------------
    box_pk := nextval('clinlims.shipping_box_seq');
    INSERT INTO clinlims.shipping_box
        (id, box_id, fhir_uuid, destination_facility_id, state, created_date, archived, sys_user_id, lastupdated)
    VALUES
        (box_pk, 'E2E-BOX-01', gen_random_uuid(), org_a::INTEGER, 'OPEN', now(), FALSE, 1, now());
    -- box_sample_item is what /items consults (getAllAssignedSampleItemIds);
    -- referral.assigned_to_box_id is what the dashboard query consults. Both
    -- are set, so the row is excluded from BOTH endpoints.
    INSERT INTO clinlims.box_sample_item
        (id, shipping_box_id, sample_item_id, added_date, sys_user_id, lastupdated)
    VALUES
        (nextval('clinlims.box_sample_item_seq'), box_pk, it_boxed::INTEGER, now(), 1, now());

    -- ---- referrals ----------------------------------------------------------
    -- Request dates are staggered so getReferralsBySampleItemId's
    -- `ORDER BY r.requestDate` is observable rather than accidental.
    INSERT INTO clinlims.referral
        (id, analysis_id, organization_id, organization_name, referral_reason_id, referral_type_id,
         referral_request_date, status, priority, lost_status, canceled, lastupdated)
    VALUES
        -- R1 -- everything populated.
        (nextval('clinlims.referral_seq'), an_1a, org_a, NULL, ref_reason, ref_type,
         TIMESTAMPTZ '2025-05-01 08:00:00+00', 'CREATED', 'STAT', FALSE, FALSE, now()),
        -- R4 -- second referral on the SAME sample item, LATER request date.
        (nextval('clinlims.referral_seq'), an_1b, org_b, NULL, ref_reason, ref_type,
         TIMESTAMPTZ '2025-05-02 08:00:00+00', 'SENT', 'ROUTINE', FALSE, FALSE, now()),
        -- R2 -- no organization, only a free-text name; priority NULL.
        (nextval('clinlims.referral_seq'), an_1b_item, NULL, 'E2E Free Text Facility', ref_reason, ref_type,
         TIMESTAMPTZ '2025-05-03 08:00:00+00', 'CREATED', NULL, FALSE, FALSE, now()),
        -- R5 -- the single-item accession.
        (nextval('clinlims.referral_seq'), an_2a, org_a, NULL, ref_reason, ref_type,
         TIMESTAMPTZ '2025-05-04 08:00:00+00', 'CREATED', 'ROUTINE', FALSE, FALSE, now()),
        -- R3 -- request date NULL and reason NULL.
        (nextval('clinlims.referral_seq'), an_2a, org_a, NULL, NULL, ref_type,
         NULL, 'CREATED', 'ROUTINE', FALSE, FALSE, now()),
        -- X1 lost
        (nextval('clinlims.referral_seq'), an_lost, org_a, NULL, ref_reason, ref_type,
         TIMESTAMPTZ '2025-05-05 08:00:00+00', 'CREATED', 'ROUTINE', TRUE, FALSE, now()),
        -- X2 canceled
        (nextval('clinlims.referral_seq'), an_cancel, org_a, NULL, ref_reason, ref_type,
         TIMESTAMPTZ '2025-05-06 08:00:00+00', 'CANCELED', 'ROUTINE', FALSE, TRUE, now()),
        -- X4 status NULL -- excluded by three-valued logic, not by an explicit rule.
        (nextval('clinlims.referral_seq'), an_nullstat, org_a, NULL, ref_reason, ref_type,
         TIMESTAMPTZ '2025-05-08 08:00:00+00', NULL, 'ROUTINE', FALSE, FALSE, now()),
        -- X5 voided sample item -- excluded from /items, PRESENT on the dashboard.
        (nextval('clinlims.referral_seq'), an_voided, org_a, NULL, ref_reason, ref_type,
         TIMESTAMPTZ '2025-05-09 08:00:00+00', 'CREATED', 'ROUTINE', FALSE, FALSE, now());
    -- X3 assigned to a box -- separate INSERT so the box FK is set.
    INSERT INTO clinlims.referral
        (id, analysis_id, organization_id, referral_reason_id, referral_type_id,
         referral_request_date, status, priority, lost_status, canceled, assigned_to_box_id, lastupdated)
    VALUES
        (nextval('clinlims.referral_seq'), an_boxed, org_a, ref_reason, ref_type,
         TIMESTAMPTZ '2025-05-07 08:00:00+00', 'CREATED', 'ROUTINE', FALSE, FALSE, box_pk, now());

    -- ---- order attachments --------------------------------------------------
    -- Real, decodable bytes. Tiny on purpose: download/view echo the content
    -- back in full and Content-Length is asserted against it.
    pdf_bytes := decode('JVBERi0xLjQKJUUyRTIK', 'base64');
    bin_bytes := decode('RTJFLUJJTkFSWQ==', 'base64');
    INSERT INTO clinlims.order_attachment
        (id, sample_id, original_file_name, file_type, file_size_bytes, file_content,
         uploaded_by, uploaded_at, is_deleted, last_updated)
    VALUES
        -- A1 typed
        (nextval('clinlims.order_attachment_seq'), s_att, 'e2e-report.pdf', 'application/pdf',
         octet_length(pdf_bytes), pdf_bytes, 1, TIMESTAMP '2025-05-04 10:00:00', FALSE, now()),
        -- A2 untyped -> "" in the list, application/octet-stream on download
        (nextval('clinlims.order_attachment_seq'), s_att, 'e2e-untyped.bin', NULL,
         octet_length(bin_bytes), bin_bytes, 1, TIMESTAMP '2025-05-04 11:00:00', FALSE, now()),
        -- A3 soft-deleted -> absent from the list, 404 on download/view
        (nextval('clinlims.order_attachment_seq'), s_att, 'e2e-deleted.pdf', 'application/pdf',
         octet_length(pdf_bytes), pdf_bytes, 1, TIMESTAMP '2025-05-04 12:00:00', TRUE, now());

    RAISE NOTICE 'shipment-attachment-e2e: seeded 10 referrals (5 visible, 5 excluded) and 3 attachments';
END $BODY$;
