-- source: liquibase liquibase/3.5.x.x/004-analysis-analyzer-id.xml::3.5.0-003-backfill-analyzer-fhir-uuid::pmanko
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.analyzer SET fhir_uuid = gen_random_uuid() WHERE fhir_uuid IS NULL;

ALTER TABLE clinlims.analyzer ALTER COLUMN  fhir_uuid SET NOT NULL;

ALTER TABLE clinlims.analyzer ADD CONSTRAINT uq_analyzer_fhir_uuid UNIQUE (fhir_uuid);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/004-analysis-analyzer-id.xml::3.5.0-003-backfill-analyzer-fhir-uuid::pmanko
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
