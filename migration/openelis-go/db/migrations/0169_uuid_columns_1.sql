-- source: liquibase liquibase/2.3.x.x/uuid_columns.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.analysis ADD IF NOT EXISTS fhir_uuid UUID;
ALTER TABLE clinlims.result ADD IF NOT EXISTS fhir_uuid UUID;
ALTER TABLE clinlims.sample_item ADD IF NOT EXISTS fhir_uuid UUID;
ALTER TABLE clinlims.organization ADD IF NOT EXISTS fhir_uuid UUID;
ALTER TABLE clinlims.patient ADD IF NOT EXISTS fhir_uuid UUID;
ALTER TABLE clinlims.sample ADD IF NOT EXISTS fhir_uuid UUID;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/uuid_columns.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
