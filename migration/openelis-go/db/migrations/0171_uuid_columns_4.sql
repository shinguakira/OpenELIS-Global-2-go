-- source: liquibase liquibase/2.3.x.x/uuid_columns.xml::4::csteele
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.provider ADD IF NOT EXISTS fhir_uuid UUID;
CREATE INDEX IF NOT EXISTS provider_fhir_uuid_index ON provider(fhir_uuid);
ALTER TABLE clinlims.referral ADD IF NOT EXISTS fhir_uuid UUID;
CREATE INDEX IF NOT EXISTS referral_fhir_uuid_index ON referral(fhir_uuid);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/uuid_columns.xml::4::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
