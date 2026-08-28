-- =============================================================================
-- Result-side data for Wave 5 (c3) — E2E fixture
-- =============================================================================
-- The c3 endpoints read results, panels, test sections and validation-pending
-- analyses. The dev dataset has NONE of that:
--
--     result                     0 rows      -> every result value is absent
--     analysis.panel_id          all NULL    -> the column the app fills when an
--                                               order is placed as a panel
--     analysis in TechnicalAcceptance
--                                0 rows      -> AccessionValidation returns nothing
--
-- So the whole wave answered with an empty envelope, and the baseline spec
-- codified that emptiness as the expectation — one test literally asserts that
-- all four WorkPlan endpoints "share ONE identical empty envelope". A port that
-- returned a hardcoded empty form would have passed it.
--
-- (analysis.test_sect_id was missing too, but that belongs to the fixtures that
-- create the analyses and is fixed there, not here.)
--
-- This file seeds ONE order, E2E-RES-01, carrying:
--   * two sample items, both with a type of sample (see the NOTE below)
--   * analyses on BOTH tests of a real panel, and analysis.panel_id set
--   * test_sect_id set, so WorkPlanByTestSection has rows
--   * numeric AND dictionary results, so result rendering is exercised both ways
--   * one analysis in TechnicalAcceptance, which is the status
--     AccessionValidation filters on
--
-- ---- NOTE: the type-less sample item is NOT copied here ---------------------
-- shipment-attachment-e2e.sql deliberately seeds one sample_item with a NULL
-- typeosamp_id, because c2 proved Java's own unassigned-sample HQL LEFT JOINs
-- type_of_sample and COALESCEs the description — it is written to tolerate that
-- state. AnalysisServiceImpl.getTestDisplayName is not, and NPEs on it, which
-- is why rest/LogbookResults?selectedTest=1 returns 500 while ?selectedTest=2
-- returns 200 with rows.
--
-- That is a JAVA DEFECT and it stays. This order keeps every item typed so it
-- can be used to verify the SUCCESS path; the 500 is pinned separately against
-- the existing type-less row. Do not "fix" either side.
--
-- ---- ID / SEQUENCE POLICY --------------------------------------------------
-- nextval everywhere, cleaned up by accession marker — the loader's
-- normalize_sequences step runs setval(seq, MAX(id) + 1) over sample,
-- sample_item and analysis, so reserved ids are not an option.
--
-- sample.status_id and sample.lastupdated are BOTH set. Java dereferences
-- status_id without a null check (QAService.nonconformingByDepricatedStatus)
-- and optimistic-locks on lastupdated; a NULL in either takes unrelated
-- endpoints from 200 to 500.
--
-- Usage (via the repo's loader, from repo root):
--   ./src/test/resources/load-test-fixtures.sh --profile=core
--
-- IDEMPOTENT: safe to re-run; every row is deleted by marker before re-insert.
-- =============================================================================

DO $BODY$
DECLARE
    target_patient  NUMERIC;
    order_status    NUMERIC;   -- status_of_sample, status_type='ORDER'
    item_status     NUMERIC;   -- status_of_sample, status_type='SAMPLE'
    tos_id          NUMERIC;
    panel_id_v      NUMERIC;
    test_x          NUMERIC;   -- first test of that panel
    test_y          NUMERIC;   -- second test of that panel
    sect_x          NUMERIC;
    sect_y          NUMERIC;
    st_notstarted   NUMERIC;
    st_techaccept   NUMERIC;
    dict_value      NUMERIC;

    s_res           NUMERIC;
    it_a            NUMERIC;
    it_b            NUMERIC;
    an_num          NUMERIC;   -- numeric result carrier
    an_dict         NUMERIC;   -- dictionary result carrier
    an_validate     NUMERIC;   -- TechnicalAcceptance, for AccessionValidation
    test_clean      NUMERIC;   -- a test in a section NO type-less item touches
    sect_clean      NUMERIC;
    an_clean        NUMERIC;
    an_banded       NUMERIC;   -- TechnicalAcceptance on a MULTI-BAND test
    an_rejected     NUMERIC;   -- Technical Rejected, for the config-gated status
    st_techreject   NUMERIC;
    ref_analysis    NUMERIC;   -- an existing referred-out analysis
    ana_table       NUMERIC;   -- reference_tables id for ANALYSIS
    banded_test     NUMERIC;   -- a test with MORE THAN ONE result_limits band
    banded_section  NUMERIC;
BEGIN
    -- ---- cleanup, children first -------------------------------------------
    DELETE FROM clinlims.result
     WHERE analysis_id IN (
        SELECT a.id FROM clinlims.analysis a
          JOIN clinlims.sample_item si ON si.id = a.sampitem_id
          JOIN clinlims.sample s ON s.id = si.samp_id
         WHERE s.accession_number = 'E2E-RES-01');
    DELETE FROM clinlims.analysis
     WHERE sampitem_id IN (
        SELECT si.id FROM clinlims.sample_item si
          JOIN clinlims.sample s ON s.id = si.samp_id
         WHERE s.accession_number = 'E2E-RES-01');
    DELETE FROM clinlims.sample_item
     WHERE samp_id IN (SELECT id FROM clinlims.sample WHERE accession_number = 'E2E-RES-01');
    DELETE FROM clinlims.sample_human
     WHERE samp_id IN (SELECT id FROM clinlims.sample WHERE accession_number = 'E2E-RES-01');
    DELETE FROM clinlims.sample WHERE accession_number = 'E2E-RES-01';

    -- ---- reference data, resolved at load time ------------------------------
    SELECT id INTO target_patient FROM clinlims.patient ORDER BY id LIMIT 1;
    SELECT id INTO tos_id  FROM clinlims.type_of_sample WHERE is_active = true ORDER BY id LIMIT 1;
    SELECT id INTO order_status FROM clinlims.status_of_sample
                                 WHERE status_type = 'ORDER'  AND name = 'Test Entered' LIMIT 1;
    SELECT id INTO item_status  FROM clinlims.status_of_sample
                                 WHERE status_type = 'SAMPLE' AND name = 'SampleEntered' LIMIT 1;

    -- Analysis statuses are resolved by the name STORED in status_of_sample,
    -- not by the AnalysisStatus enum constant: StatusService.addToAnalysisMap
    -- matches literally, so NotStarted is the row named 'Not Tested' and
    -- TechnicalAcceptance is 'Technical Acceptance'.
    SELECT id INTO st_notstarted FROM clinlims.status_of_sample
                                  WHERE status_type = 'ANALYSIS' AND name = 'Not Tested' LIMIT 1;
    SELECT id INTO st_techaccept FROM clinlims.status_of_sample
                                  WHERE status_type = 'ANALYSIS' AND name = 'Technical Acceptance' LIMIT 1;
    SELECT id INTO st_techreject FROM clinlims.status_of_sample
                                  WHERE status_type = 'ANALYSIS' AND name = 'Technical Rejected' LIMIT 1;

    -- A test carrying MORE THAN ONE result_limits row. result_limits is keyed
    -- by test AND by age band (min_age/max_age in DAYS) and optionally gender,
    -- so a naive join on test_id alone multiplies the analysis by its band
    -- count. Every analysis seeded before this sat on a SINGLE-band test, which
    -- is why that mistake still produced a byte-identical response.
    SELECT rl.test_id, t.test_section_id INTO banded_test, banded_section
      FROM clinlims.result_limits rl
      JOIN clinlims.test t ON t.id = rl.test_id
     GROUP BY rl.test_id, t.test_section_id
    HAVING count(*) > 1
     ORDER BY rl.test_id LIMIT 1;

    SELECT id INTO ana_table FROM clinlims.reference_tables WHERE name = 'ANALYSIS' LIMIT 1;

    -- A panel with at least two tests, and those two tests.
    --
    -- WorkPlanByPanel does NOT filter on analysis.panel_id — measured: it reads
    -- panel_item for the panel and then runs getAllAnalysisByTestAndStatus once
    -- PER MEMBER TEST, concatenating the results. panel_id=1 returns 16 rows =
    -- test 1's 12 plus test 2's 4, and no analysis outside those tests appears
    -- however its panel_id is set.
    --
    -- panel_id is still written here because SamplePatientEntryServiceImpl and
    -- SampleEditServiceImpl both call analysis.setPanel() when an order is
    -- placed as a panel, so a panel order with it NULL is data the app cannot
    -- produce — the same class of unfaithful fixture as a NULL test_sect_id.
    -- It is simply not what makes this endpoint answer.
    SELECT pi.panel_id INTO panel_id_v
      FROM clinlims.panel_item pi
     GROUP BY pi.panel_id HAVING count(*) >= 2
     ORDER BY pi.panel_id LIMIT 1;
    SELECT pi.test_id INTO test_x FROM clinlims.panel_item pi
     WHERE pi.panel_id = panel_id_v ORDER BY pi.test_id LIMIT 1;
    SELECT pi.test_id INTO test_y FROM clinlims.panel_item pi
     WHERE pi.panel_id = panel_id_v AND pi.test_id <> test_x ORDER BY pi.test_id LIMIT 1;
    SELECT test_section_id INTO sect_x FROM clinlims.test WHERE id = test_x;
    SELECT test_section_id INTO sect_y FROM clinlims.test WHERE id = test_y;

    -- A dictionary row, for the non-numeric result branch.
    SELECT id INTO dict_value FROM clinlims.dictionary
     WHERE dict_entry IS NOT NULL ORDER BY id LIMIT 1;

    -- A test whose section contains NO analysis sitting on a type-less sample
    -- item. Without this, rest/WorkPlanByTestSection can only ever be observed
    -- 500ing: the one section that has analyses (Biochemistry) also has the
    -- deliberately type-less item from shipment-attachment-e2e.sql, and
    -- AnalysisServiceImpl.getTestDisplayName NPEs on it.
    --
    -- The 500 is a JAVA DEFECT and stays pinned on that section. This gives the
    -- SUCCESS path a section of its own so both can be verified.
    SELECT t.id, t.test_section_id INTO test_clean, sect_clean
      FROM clinlims.test t
     WHERE t.test_section_id IS NOT NULL
       AND t.is_active = 'Y'
       AND t.test_section_id NOT IN (
           SELECT DISTINCT t2.test_section_id
             FROM clinlims.analysis a
             JOIN clinlims.sample_item si ON si.id = a.sampitem_id
             JOIN clinlims.test t2 ON t2.id = a.test_id
            WHERE si.typeosamp_id IS NULL AND t2.test_section_id IS NOT NULL)
     ORDER BY t.test_section_id, t.id LIMIT 1;

    IF target_patient IS NULL OR tos_id IS NULL OR order_status IS NULL
       OR item_status IS NULL OR panel_id_v IS NULL OR test_x IS NULL OR test_y IS NULL
       OR st_notstarted IS NULL OR st_techaccept IS NULL THEN
        RAISE NOTICE 'result-reads-e2e: prerequisites missing; nothing seeded.';
        RETURN;
    END IF;

    -- ---- the order ----------------------------------------------------------
    s_res := nextval('clinlims.sample_seq');
    INSERT INTO clinlims.sample
        (id, accession_number, entered_date, received_date, collection_date,
         order_priority, status_id, lastupdated, is_confirmation)
    VALUES
        (s_res, 'E2E-RES-01', now(), TIMESTAMP '2025-08-01 09:00:00',
         TIMESTAMP '2025-08-01 08:45:00', 'ROUTINE', order_status, now(), false);

    INSERT INTO clinlims.sample_human (id, samp_id, patient_id)
    VALUES (nextval('clinlims.sample_human_seq'), s_res, target_patient);

    -- Both items are TYPED on purpose — see the note in the header.
    it_a := nextval('clinlims.sample_item_seq');
    it_b := nextval('clinlims.sample_item_seq');
    INSERT INTO clinlims.sample_item
        (id, samp_id, sort_order, typeosamp_id, status_id, collection_date, voided, rejected, lastupdated)
    VALUES
        (it_a, s_res, 1, tos_id, item_status, TIMESTAMP '2025-08-01 08:45:00', FALSE, FALSE, now()),
        (it_b, s_res, 2, tos_id, item_status, TIMESTAMP '2025-08-01 08:45:00', FALSE, FALSE, now());

    -- ---- analyses ------------------------------------------------------------
    -- panel_id AND test_sect_id are both set: the two WorkPlan endpoints that
    -- returned nothing filter on exactly these two columns.
    an_num      := nextval('clinlims.analysis_seq');
    an_dict     := nextval('clinlims.analysis_seq');
    an_validate := nextval('clinlims.analysis_seq');
    an_clean    := nextval('clinlims.analysis_seq');
    INSERT INTO clinlims.analysis
        (id, sampitem_id, test_id, test_sect_id, panel_id, status_id,
         analysis_type, entry_date, is_reportable, revision, lastupdated)
    VALUES
        (an_num,      it_a, test_x, sect_x, panel_id_v, st_notstarted, 'MANUAL', now(), 'Y', 0, now()),
        (an_dict,     it_a, test_y, sect_y, panel_id_v, st_notstarted, 'MANUAL', now(), 'Y', 0, now()),
        -- Technical Acceptance = awaiting biologist validation. This is the one
        -- status AccessionValidation looks for, and nothing in the dataset was
        -- in it, so that endpoint could only ever answer with an empty list.
        (an_validate, it_b, test_x, sect_x, panel_id_v, st_techaccept, 'MANUAL', now(), 'Y', 0, now());

    -- The clean-section analysis. No panel_id: WorkPlanByPanel must not pick it
    -- up, which is what proves that endpoint filters on the column rather than
    -- returning everything.
    IF test_clean IS NOT NULL THEN
        INSERT INTO clinlims.analysis
            (id, sampitem_id, test_id, test_sect_id, status_id,
             analysis_type, entry_date, is_reportable, revision, lastupdated)
        VALUES
            (an_clean, it_b, test_clean, sect_clean, st_notstarted, 'MANUAL', now(), 'Y', 0, now());
    END IF;

    -- ---- results -------------------------------------------------------------
    -- result was entirely empty, so no c3 endpoint had a value to render and a
    -- port that emitted "" everywhere agreed with Java on every field.
    --
    -- Two result_types, because they render differently:
    --   'N' numeric  -> value is the number as text, min/max_normal populated
    --   'D' dictionary -> value is a DICTIONARY ID, resolved to its name for
    --                     display, exactly like observation_history's non-literal
    --                     branch in c2
    INSERT INTO clinlims.result
        (id, analysis_id, sort_order, is_reportable, result_type, value,
         min_normal, max_normal, significant_digits, lastupdated)
    VALUES
        (nextval('clinlims.result_seq'), an_num,  1, 'Y', 'N', '42.5', 10, 50, 1, now()),
        (nextval('clinlims.result_seq'), an_validate, 1, 'Y', 'N', '7.25', 1, 9, 2, now());

    IF dict_value IS NOT NULL THEN
        INSERT INTO clinlims.result
            (id, analysis_id, sort_order, is_reportable, result_type, value, lastupdated)
        VALUES (nextval('clinlims.result_seq'), an_dict, 1, 'Y', 'D', dict_value::text, now());
    END IF;

    -- ---- analyses the review found unreachable ------------------------------

    -- (a) TechnicalAcceptance on a MULTI-BAND test, so AccessionValidation has
    --     to pick ONE result_limits row rather than joining them all.
    IF banded_test IS NOT NULL AND st_techaccept IS NOT NULL THEN
        an_banded := nextval('clinlims.analysis_seq');
        INSERT INTO clinlims.analysis
            (id, sampitem_id, test_id, test_sect_id, status_id,
             analysis_type, entry_date, is_reportable, revision, lastupdated)
        VALUES
            (an_banded, it_b, banded_test, banded_section, st_techaccept,
             'MANUAL', now(), 'Y', 0, now());
        INSERT INTO clinlims.result
            (id, analysis_id, sort_order, is_reportable, result_type, value,
             min_normal, max_normal, significant_digits, lastupdated)
        VALUES (nextval('clinlims.result_seq'), an_banded, 1, 'Y', 'N', '6.5', 4, 10, 2, now());
    END IF;

    -- (b) Technical Rejected. AccessionValidation adds this status ONLY when
    --     site_information validateTechnicalRejection is true, and NOTHING in
    --     the dataset was in it — so neither branch of that condition was
    --     observable and a port that ignored the setting matched anyway.
    IF st_techreject IS NOT NULL THEN
        an_rejected := nextval('clinlims.analysis_seq');
        INSERT INTO clinlims.analysis
            (id, sampitem_id, test_id, test_sect_id, status_id,
             analysis_type, entry_date, is_reportable, revision, lastupdated)
        VALUES
            (an_rejected, it_a, test_y, sect_y, st_techreject,
             'MANUAL', now(), 'Y', 0, now());
        INSERT INTO clinlims.result
            (id, analysis_id, sort_order, is_reportable, result_type, value,
             min_normal, max_normal, significant_digits, lastupdated)
        VALUES (nextval('clinlims.result_seq'), an_rejected, 1, 'Y', 'N', '99', 1, 9, 0, now());
    END IF;

    -- (c) started_date on one analysis. AccessionValidation's third search
    --     branch is `date`, which filters getAnalysisStartedOn — and no
    --     analysis in the dataset had a started_date, so that branch could
    --     only ever return nothing and its absence from the port was invisible.
    UPDATE clinlims.analysis SET started_date = DATE '2025-08-01'
     WHERE id = an_validate;

    -- (d) A result and a NOTE on an already-referred analysis. Java's
    --     convertToDisplayItem fills referralResultsDisplay, resultDate and
    --     notes only when those exist; every seeded referral had neither, so
    --     three fields were unreachable.
    SELECT a.id INTO ref_analysis
      FROM clinlims.referral r
      JOIN clinlims.analysis a ON a.id = r.analysis_id
      JOIN clinlims.sample_item si ON si.id = a.sampitem_id
      JOIN clinlims.sample s ON s.id = si.samp_id
     WHERE s.accession_number = 'E2E-REF-01'
     ORDER BY a.id LIMIT 1;
    IF ref_analysis IS NOT NULL THEN
        DELETE FROM clinlims.result WHERE analysis_id = ref_analysis;
        INSERT INTO clinlims.result
            (id, analysis_id, sort_order, is_reportable, result_type, value,
             min_normal, max_normal, significant_digits, lastupdated)
        VALUES (nextval('clinlims.result_seq'), ref_analysis, 1, 'Y', 'N', '13.75', 1, 9, 2, now());
        UPDATE clinlims.analysis SET completed_date = DATE '2025-05-20' WHERE id = ref_analysis;
        IF ana_table IS NOT NULL THEN
            DELETE FROM clinlims.note
             WHERE reference_table = ana_table AND reference_id = ref_analysis
               AND subject = 'E2E-REFERRAL-NOTE';
            INSERT INTO clinlims.note
                (id, sys_user_id, reference_id, reference_table, note_type, subject, text, lastupdated)
            VALUES (nextval('clinlims.note_seq'), 1, ref_analysis, ana_table, 'I',
                    'E2E-REFERRAL-NOTE', 'E2E referral note text', now());
        END IF;
    END IF;

    RAISE NOTICE 'result-reads-e2e: seeded E2E-RES-01 (sample %) — panel %, tests %/%, sections %/%, clean section % (test %), results',
        s_res, panel_id_v, test_x, test_y, sect_x, sect_y, sect_clean, test_clean;
END $BODY$;
