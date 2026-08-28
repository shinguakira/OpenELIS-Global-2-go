-- =============================================================================
-- Fully-populated orders for rest/order/search — E2E fixture
-- =============================================================================
-- `buildSampleOrderItems` emits far more than the six keys the dev dataset
-- happens to exercise. Everything below is conditional in Java, and NONE of the
-- conditions are met by any pre-existing sample:
--
--     providerPersonId / providerFirstName / providerLastName /
--     providerWorkPhone / providerEmail / providerFax
--                                     <- sample_human.provider_id -> person
--     referringSiteId / referringSiteName / referringSiteCode
--     referringSiteDepartmentId / referringSiteDepartmentName
--                                     <- sample_requester, split by ORG TYPE
--     program / programId             <- observation_history AND program_sample
--     paymentOptionSelection          <- observation_history paymentStatus
--     billingReferenceNumber          <- observation_history billingRefNumber
--     testLocationCode                <- observation_history testLocationCode
--     otherLocationCode               <- observation_history testLocationCodeOther
--     requestDate                     <- observation_history requestDate
--     nextVisitDate                   <- observation_history nextVisitDate
--     provisionalClinicalDiagnosis    <- observation_history
--     priority                        <- sample.order_priority
--
-- With none of them present the Go port returned a six-key object and the
-- parity suite was green, because the fixture had nothing to disagree about.
-- That is the "green run certifying a parity that does not exist" failure the
-- migration guide warns about, and it was found by review rather than by a
-- test — the endpoint was declared finished with most of its builder unported.
--
-- ---- THREE ORDERS, BECAUSE PROGRAM RESOLUTION HAS THREE BRANCHES ------------
-- The program keys are the subtlest part of the builder. Java does NOT simply
-- read program_sample.program_id:
--
--   (a) program observation present, name is NOT pathology/cytology/
--       immunohistochemistry
--          -> getProgrammeSampleBySample picks entity class ProgramSample,
--             queries program_sample, finds the row
--          -> programId comes from program_sample.program_id
--          -> program is the OBSERVATION value
--
--   (b) program observation present and its name contains "cytology",
--       "pathology" or "immunohistochemistry"
--          -> the entity class becomes CytologySample / PathologySample /
--             ImmunohistochemistrySample. Those are @Inheritance(TABLE_PER_CLASS)
--             and live in their OWN tables, so the program_sample row is
--             invisible and the query returns null
--          -> Java falls back to scanning programService.getAll() for a program
--             whose NAME equals the observation value
--          -> programId is the id of the NAMED program, which need not be the
--             one program_sample points at
--
--   (c) NO program observation
--          -> the else branch reads program_sample via a LIKE on the accession
--             number and takes BOTH keys from it: programId = program.id and
--             program = program.NAME (not an observation value at all)
--
-- E2E-FULL-01 / -02 / -03 are (a) / (b) / (c) respectively. -02 deliberately
-- points program_sample at a DIFFERENT program from the one its observation
-- names, which is the only way branch (b) is distinguishable from (a): a port
-- that just reads program_sample.program_id returns the wrong id there and the
-- right one everywhere else.
--
-- -02 additionally carries a value_type='D' observation, whose value is a
-- DICTIONARY ID rather than text: getValueForSample renders those through
-- dictionary.getLocalizedName() while getRawValueForSample returns the id, and
-- with every observation stored as LITERAL the difference was invisible.
--
-- -02 and -03 also carry NO provider and NO requester, and -03 an explicit NULL
-- order_priority, so together they invert every conditional key -01 emits.
--
-- ---- WHAT IS DELIBERATELY *NOT* SEEDED --------------------------------------
--   additionalQuestions  -- a FHIR QuestionnaireResponse fetched from the FHIR
--                           server, not a DB row. Out of reach of a SQL fixture
--                           and out of scope for a read-only port.
--   environmentalFields  -- the env* observation types belong to the
--                           environmental workflow, which is its own feature.
--                           E2E-REF-* / E2E-EDIT-* stay clinical; a dedicated
--                           environmental order belongs with that work.
--   Both are called out here so their absence is a stated limit rather than an
--   accident, and so the next person knows the map is still not exhaustive.
--
-- ---- ID / SEQUENCE POLICY --------------------------------------------------
-- nextval everywhere, cleaned up by marker. Reserved ids are not an option for
-- sample / sample_item / analysis / sample_human: the loader's
-- normalize_sequences step runs setval(seq, MAX(id) + 1) over those tables.
-- e2e-foundational-data.sql owns the 9000xxx block, so nothing here goes near
-- it.
--
-- lastupdated is set on the sample rows. Hibernate optimistic-locks Sample on
-- that column, and a NULL there makes any dirty-check flush fail with
-- StaleStateException — which once took rest/order/dashboard from 200 to 500
-- for the whole table.
--
-- Usage (via the repo's loader, from repo root):
--   ./src/test/resources/load-test-fixtures.sh --profile=core
--
-- IDEMPOTENT: safe to re-run; every row is deleted by marker before re-insert.
-- =============================================================================

DO $BODY$
DECLARE
    order_status    NUMERIC;   -- status_of_sample, status_type='ORDER'
    target_patient  NUMERIC;
    v_person        NUMERIC;
    v_provider      NUMERIC;
    site_org        NUMERIC;
    dept_org        NUMERIC;
    dept_type       NUMERIC;
    rt_provider     NUMERIC;   -- requester_type named 'provider' (= PERSON)
    rt_org          NUMERIC;   -- requester_type named 'organization'
    dict_value      NUMERIC;   -- a dictionary row, for the non-LITERAL branch
    tos_id          NUMERIC;
    samp_status     NUMERIC;
    prog_plain      NUMERIC;   -- a program whose name trips NO subclass rule
    prog_plain_name TEXT;
    prog_sub        NUMERIC;   -- a program whose name DOES trip one
    prog_sub_name   TEXT;
    s_full          NUMERIC;
    s_sub           NUMERIC;
    s_noobs         NUMERIC;

    -- observation_history_type ids, resolved by name below.
    t_payment       NUMERIC;
    t_billing       NUMERIC;
    t_testloc       NUMERIC;
    t_testloc_other NUMERIC;
    t_request       NUMERIC;
    t_nextvisit     NUMERIC;
    t_diagnosis     NUMERIC;
    t_program       NUMERIC;
BEGIN
    -- sample.status_id is the ORDER-level status (status_type='ORDER'), NOT
    -- the SAMPLE-level one used by sample_item. Every stock sample carries it,
    -- and Java dereferences it without a null check, so leaving it NULL breaks
    -- unrelated endpoints (WorkPlanByTest 500s on the resulting NPE).
    SELECT id INTO order_status FROM clinlims.status_of_sample
     WHERE status_type = 'ORDER' AND name = 'Test Entered' LIMIT 1;
    -- ---- cleanup, children first -------------------------------------------
    DELETE FROM clinlims.observation_history
     WHERE sample_id IN (SELECT id FROM clinlims.sample WHERE accession_number LIKE 'E2E-FULL-%');
    DELETE FROM clinlims.program_sample
     WHERE sample_id IN (SELECT id FROM clinlims.sample WHERE accession_number LIKE 'E2E-FULL-%');
    DELETE FROM clinlims.sample_requester
     WHERE sample_id IN (SELECT id FROM clinlims.sample WHERE accession_number LIKE 'E2E-FULL-%');
    DELETE FROM clinlims.analysis
     WHERE sampitem_id IN (
        SELECT si.id FROM clinlims.sample_item si
          JOIN clinlims.sample s ON s.id = si.samp_id
         WHERE s.accession_number LIKE 'E2E-FULL-%');
    DELETE FROM clinlims.sample_item
     WHERE samp_id IN (SELECT id FROM clinlims.sample WHERE accession_number LIKE 'E2E-FULL-%');
    DELETE FROM clinlims.sample_human
     WHERE samp_id IN (SELECT id FROM clinlims.sample WHERE accession_number LIKE 'E2E-FULL-%');
    DELETE FROM clinlims.sample WHERE accession_number LIKE 'E2E-FULL-%';

    -- The dedicated provider and the department organization are owned by this
    -- fixture, so they are dropped and rebuilt too. Deleting the provider first
    -- keeps the person's FK satisfied.
    DELETE FROM clinlims.provider
     WHERE person_id IN (SELECT id FROM clinlims.person WHERE email = 'e2e.full.provider@example.test');
    DELETE FROM clinlims.person WHERE email = 'e2e.full.provider@example.test';
    -- 部門組織は DELETE しない。毎回作り直すと nextval で id が動き、この
    -- 組織を指している他の fixture の行（referral.organization_id など）が
    -- 宙に浮く。実際それで shipment fixture の可視 referral が 1件減った。
    -- 下で「あれば再利用、無ければ作成」する。

    -- ---- reference data, resolved at load time ------------------------------
    SELECT id INTO target_patient FROM clinlims.patient ORDER BY id LIMIT 1;
    SELECT id INTO tos_id        FROM clinlims.type_of_sample WHERE is_active = true ORDER BY id LIMIT 1;
    SELECT id INTO samp_status   FROM clinlims.status_of_sample
                                  WHERE status_type = 'SAMPLE' AND name = 'SampleEntered' LIMIT 1;

    -- The site/department split is decided by the ORGANIZATION TYPE, not by
    -- sample_requester.requester_type_id — that column only separates
    -- organization(1) from provider(2). RequesterService asks
    -- getOrganizationRequester twice, once for the type named "referring clinic"
    -- and once for "dept", so the two organizations must carry those types.
    SELECT o.id INTO site_org FROM clinlims.organization o
      JOIN clinlims.organization_organization_type oot ON oot.org_id = o.id
      JOIN clinlims.organization_type t ON t.id = oot.org_type_id
     WHERE t.short_name = 'referring clinic' ORDER BY o.id LIMIT 1;

    SELECT id INTO dept_type FROM clinlims.organization_type WHERE short_name = 'dept' LIMIT 1;

    -- requester_type is NOT what it looks like. SampleServiceImpl resolves
    -- PERSON_REQUESTER_TYPE_ID from the row named 'provider', and
    -- getPersonRequester then reads that row's requester_id as a PERSON id.
    -- So 'provider' means "requester_id is a person", and 'organization' means
    -- "requester_id is an organization".
    SELECT id INTO rt_provider FROM clinlims.requester_type WHERE requester_type = 'provider' LIMIT 1;
    SELECT id INTO rt_org      FROM clinlims.requester_type WHERE requester_type = 'organization' LIMIT 1;

    -- A dictionary row for the observation whose value_type is NOT 'L'.
    -- getValueForSample renders such a value as dictionary.getDataForId(value)
    -- .getLocalizedName() instead of returning it verbatim; with every seeded
    -- observation stored as LITERAL that branch was unreachable and a port
    -- could return the raw dictionary ID and still pass.
    SELECT id INTO dict_value FROM clinlims.dictionary WHERE dict_entry IS NOT NULL ORDER BY id LIMIT 1;

    SELECT id INTO t_payment       FROM clinlims.observation_history_type WHERE type_name = 'paymentStatus';
    SELECT id INTO t_billing       FROM clinlims.observation_history_type WHERE type_name = 'billingRefNumber';
    SELECT id INTO t_testloc       FROM clinlims.observation_history_type WHERE type_name = 'testLocationCode';
    SELECT id INTO t_testloc_other FROM clinlims.observation_history_type WHERE type_name = 'testLocationCodeOther';
    SELECT id INTO t_request       FROM clinlims.observation_history_type WHERE type_name = 'requestDate';
    SELECT id INTO t_nextvisit     FROM clinlims.observation_history_type WHERE type_name = 'nextVisitDate';
    SELECT id INTO t_diagnosis     FROM clinlims.observation_history_type WHERE type_name = 'provisionalClinicalDiagnosis';
    SELECT id INTO t_program       FROM clinlims.observation_history_type WHERE type_name = 'program';

    -- Branch (a) needs a program whose name trips none of the three subclass
    -- rules; branch (b) needs one that trips a rule AND is a different row.
    SELECT id, name INTO prog_plain, prog_plain_name FROM clinlims.program
     WHERE lower(name) NOT LIKE '%pathology%'
       AND lower(name) NOT LIKE '%cytology%'
       AND lower(name) NOT LIKE '%immunohistochemistry%'
     ORDER BY id LIMIT 1;
    SELECT id, name INTO prog_sub, prog_sub_name FROM clinlims.program
     WHERE lower(name) LIKE '%pathology%'
        OR lower(name) LIKE '%cytology%'
        OR lower(name) LIKE '%immunohistochemistry%'
     ORDER BY id LIMIT 1;

    IF target_patient IS NULL OR tos_id IS NULL OR samp_status IS NULL
       OR site_org IS NULL OR dept_type IS NULL OR prog_plain IS NULL THEN
        RAISE NOTICE 'order-search-full-e2e: prerequisites missing; nothing seeded.';
        RETURN;
    END IF;

    -- ---- a provider whose person has EVERY contact column populated ---------
    -- The six provider keys sit inside one guard, but work_phone, email and fax
    -- are put RAW: an empty column drops just that key. Every existing provider
    -- person in this database has all three empty, so those three keys were
    -- unreachable. This fixture owns its provider rather than mutating a shared
    -- one.
    v_person := nextval('clinlims.person_seq');
    INSERT INTO clinlims.person
        (id, first_name, last_name, work_phone, email, fax, lastupdated)
    VALUES
        (v_person, 'E2EFull', 'Ordering-Provider', '+81-3-0000-0001',
         'e2e.full.provider@example.test', '+81-3-0000-0002', now());

    v_provider := nextval('clinlims.provider_seq');
    INSERT INTO clinlims.provider (id, person_id, active, lastupdated)
    VALUES (v_provider, v_person, TRUE, now());

    -- ---- a department-typed organization -----------------------------------
    -- organization_type 'dept' exists in the schema but NO organization carried
    -- it, so referringSiteDepartmentId/Name could never be emitted and Java's
    -- "promote the department into the site slot" fallback was the only path
    -- ever taken. Both branches are reachable now.
    --
    -- fhir_uuid is NOT optional here even though the column is nullable: every
    -- other organization in the database has one, organization-list emits the
    -- key only when it is non-null, and b2-organization.spec.ts requires it on
    -- every row. A NULL here fails that spec — which is the fixture's bug, not
    -- the spec's. The value is fixed rather than gen_random_uuid() so re-runs
    -- stay idempotent; 'dea7' keeps it hex while still reading as a marker.
    SELECT id INTO dept_org FROM clinlims.organization WHERE short_name = 'E2E-DEPT' LIMIT 1;
    IF dept_org IS NULL THEN
        dept_org := nextval('clinlims.organization_seq');
        INSERT INTO clinlims.organization
            (id, name, short_name, mls_sentinel_lab_flag, is_active, fhir_uuid, lastupdated)
        VALUES (dept_org, 'E2E Full Order Department', 'E2E-DEPT', 'N', 'Y',
                'e2e0dea7-0000-4000-8000-000000000001'::uuid, now());
    END IF;
    -- 型の紐付けだけは冪等に貼り直す（組織行そのものは残す）。
    DELETE FROM clinlims.organization_organization_type
     WHERE org_id = dept_org AND org_type_id = dept_type;
    INSERT INTO clinlims.organization_organization_type (org_id, org_type_id)
    VALUES (dept_org, dept_type);

    -- =========================================================================
    -- E2E-FULL-01 — branch (a): every key at once
    -- =========================================================================
    -- order_priority is set so the `priority` key appears; -02 and -03 leave it
    -- NULL, which is what makes the key's absence provable.
    s_full := nextval('clinlims.sample_seq');
    INSERT INTO clinlims.sample
        (id, accession_number, entered_date, received_date, collection_date,
         order_priority, lastupdated, is_confirmation, status_id)
    VALUES
        (s_full, 'E2E-FULL-01', now(), TIMESTAMP '2025-07-01 08:30:00',
         TIMESTAMP '2025-07-01 08:00:00', 'STAT', now(), false, order_status);

    -- provider_id on sample_human is what getProviderForSample reads. Every
    -- other sample_human row in the dataset leaves it NULL, which is why the
    -- six provider* keys never appeared.
    INSERT INTO clinlims.sample_human (id, samp_id, patient_id, provider_id)
    VALUES (nextval('clinlims.sample_human_seq'), s_full, target_patient, v_provider);

    INSERT INTO clinlims.sample_item
        (id, samp_id, sort_order, typeosamp_id, status_id, collection_date, voided, rejected, lastupdated)
    VALUES (nextval('clinlims.sample_item_seq'), s_full, 1, tos_id, samp_status,
            TIMESTAMP '2025-07-01 08:00:00', FALSE, FALSE, now());

    -- requester_type_id 1 = organization (2 = provider). BOTH rows here are
    -- organizations; what distinguishes site from department is the org type.
    INSERT INTO clinlims.sample_requester (id, sample_id, requester_id, requester_type_id, lastupdated)
    SELECT nextval('clinlims.sample_requester_seq'), s_full, o, rt_org, now()
      FROM (VALUES (site_org), (dept_org)) AS v(o)
     WHERE o IS NOT NULL;

    -- The PERSON requester, which is what SampleEdit's provider* keys come
    -- from. order/search sources the provider from sample_human.provider_id
    -- instead, so the two endpoints read different tables and BOTH rows are
    -- needed to exercise both. requester_id is the PERSON id, not the
    -- provider id.
    IF rt_provider IS NOT NULL THEN
        INSERT INTO clinlims.sample_requester (id, sample_id, requester_id, requester_type_id, lastupdated)
        VALUES (nextval('clinlims.sample_requester_seq'), s_full, v_person, rt_provider, now());
    END IF;

    -- Branch (a): the observation names the SAME program program_sample points
    -- at, so both sources agree and programId comes from program_sample.
    INSERT INTO clinlims.program_sample (id, sample_id, program_id, last_updated)
    VALUES (nextval('clinlims.program_sample_seq'), s_full, prog_plain, now());

    -- One row per conditional key. Values are distinct strings so a port that
    -- reads the wrong observation type is caught by the VALUE, not merely by
    -- the key being present.
    INSERT INTO clinlims.observation_history
        (id, sample_id, patient_id, observation_history_type_id, value_type, value, lastupdated)
    SELECT nextval('clinlims.observation_history_seq'), s_full, target_patient, t.id, 'L', t.val, now()
      FROM (VALUES
              (t_payment,       'E2E-PAYMENT-STATUS'),
              (t_billing,       'E2E-BILLING-REF-99'),
              (t_testloc,       'E2E-TEST-LOCATION'),
              (t_testloc_other, 'E2E-OTHER-LOCATION'),
              (t_request,       '01/07/2025'),
              (t_nextvisit,     '01/08/2025'),
              (t_diagnosis,     'E2E provisional diagnosis'),
              (t_program,       prog_plain_name)
           ) AS t(id, val)
     WHERE t.id IS NOT NULL AND t.val IS NOT NULL;

    -- =========================================================================
    -- E2E-FULL-02 — branch (b): the subclass-table fallback
    -- =========================================================================
    -- The observation names a program whose name routes the lookup to a
    -- TABLE_PER_CLASS subclass table, which holds no row for this sample. Java
    -- therefore ignores program_sample entirely and resolves the id by NAME.
    -- program_sample deliberately points somewhere else, so the two answers
    -- differ and a port that reads program_sample.program_id is caught.
    --
    -- Skipped when the deployment has no pathology/cytology/immunohistochemistry
    -- program at all; the loader still succeeds and the branch is simply
    -- untested there.
    IF prog_sub IS NOT NULL AND prog_sub <> prog_plain THEN
        s_sub := nextval('clinlims.sample_seq');
        INSERT INTO clinlims.sample
            (id, accession_number, entered_date, received_date, collection_date,
             lastupdated, is_confirmation, status_id)
        VALUES
            (s_sub, 'E2E-FULL-02', now(), TIMESTAMP '2025-07-02 08:30:00',
             TIMESTAMP '2025-07-02 08:00:00', now(), false, order_status);
        INSERT INTO clinlims.sample_human (id, samp_id, patient_id)
        VALUES (nextval('clinlims.sample_human_seq'), s_sub, target_patient);
        INSERT INTO clinlims.sample_item
            (id, samp_id, sort_order, typeosamp_id, status_id, collection_date, voided, rejected, lastupdated)
        VALUES (nextval('clinlims.sample_item_seq'), s_sub, 1, tos_id, samp_status,
                TIMESTAMP '2025-07-02 08:00:00', FALSE, FALSE, now());

        INSERT INTO clinlims.program_sample (id, sample_id, program_id, last_updated)
        VALUES (nextval('clinlims.program_sample_seq'), s_sub, prog_plain, now());

        INSERT INTO clinlims.observation_history
            (id, sample_id, patient_id, observation_history_type_id, value_type, value, lastupdated)
        VALUES (nextval('clinlims.observation_history_seq'), s_sub, target_patient,
                t_program, 'L', prog_sub_name, now());

        -- value_type 'D': the value is a DICTIONARY ID, and getValueForSample
        -- renders the dictionary row's localized name. Stored on -02 so -01
        -- stays entirely LITERAL and the two branches are separable.
        IF dict_value IS NOT NULL AND t_diagnosis IS NOT NULL THEN
            INSERT INTO clinlims.observation_history
                (id, sample_id, patient_id, observation_history_type_id, value_type, value, lastupdated)
            VALUES (nextval('clinlims.observation_history_seq'), s_sub, target_patient,
                    t_diagnosis, 'D', dict_value::text, now());
        END IF;
    END IF;

    -- =========================================================================
    -- E2E-FULL-03 — branch (c): no program observation at all
    -- =========================================================================
    -- With no observation, `program` is NOT an observation value: Java reads
    -- program_sample by a LIKE on the accession number and takes the program's
    -- own NAME. A port that only ever sources `program` from observation history
    -- omits the key here.
    s_noobs := nextval('clinlims.sample_seq');
    -- order_priority is written as an EXPLICIT NULL. The column is nullable but
    -- carries DEFAULT 'ROUTINE', so omitting it (as -01 and -02 do) stores
    -- ROUTINE and the `priority` key is always emitted. Only a deliberate NULL
    -- reaches Java's `if (sample.getPriority() != null)` guard, which is what
    -- makes the key's ABSENCE observable at all.
    INSERT INTO clinlims.sample
        (id, accession_number, entered_date, received_date, collection_date,
         order_priority, lastupdated, is_confirmation, status_id)
    VALUES
        (s_noobs, 'E2E-FULL-03', now(), TIMESTAMP '2025-07-03 08:30:00',
         TIMESTAMP '2025-07-03 08:00:00', NULL, now(), false, order_status);
    INSERT INTO clinlims.sample_human (id, samp_id, patient_id)
    VALUES (nextval('clinlims.sample_human_seq'), s_noobs, target_patient);
    INSERT INTO clinlims.sample_item
        (id, samp_id, sort_order, typeosamp_id, status_id, collection_date, voided, rejected, lastupdated)
    VALUES (nextval('clinlims.sample_item_seq'), s_noobs, 1, tos_id, samp_status,
            TIMESTAMP '2025-07-03 08:00:00', FALSE, FALSE, now());
    INSERT INTO clinlims.program_sample (id, sample_id, program_id, last_updated)
    VALUES (nextval('clinlims.program_sample_seq'), s_noobs, prog_plain, now());

    RAISE NOTICE 'order-search-full-e2e: seeded E2E-FULL-01/02/03 (samples %, %, %), provider %, department org %',
        s_full, s_sub, s_noobs, v_provider, dept_org;
END $BODY$;
