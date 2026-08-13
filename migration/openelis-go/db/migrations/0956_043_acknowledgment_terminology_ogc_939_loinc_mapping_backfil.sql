-- source: liquibase liquibase/3.5.x.x/043-acknowledgment-terminology.xml::OGC-939-loinc-mapping-backfill::OGC
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.test_terminology_mapping
                (id, test_id, source, code, relationship, is_active, lastupdated)
            SELECT gen_random_uuid()::varchar, t.id, 'LOINC', t.loinc, 'EQUIVALENT', 'Y', now()
              FROM clinlims.test t
             WHERE t.loinc IS NOT NULL AND length(trim(t.loinc)) > 0
               AND NOT EXISTS (SELECT 1 FROM clinlims.test_terminology_mapping m
                                WHERE m.test_id = t.id AND m.source = 'LOINC' AND m.code = t.loinc) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/043-acknowledgment-terminology.xml::OGC-939-loinc-mapping-backfill::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
