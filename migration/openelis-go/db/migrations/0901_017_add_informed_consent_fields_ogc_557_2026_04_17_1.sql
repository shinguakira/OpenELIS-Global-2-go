-- source: liquibase liquibase/3.5.x.x/017-add-informed-consent-fields.xml::OGC-557-2026-04-17-1::herbertYiga
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.sample ADD IF NOT EXISTS consent_provided BOOLEAN;
ALTER TABLE clinlims.sample ADD IF NOT EXISTS consent_reference_no VARCHAR(100);
ALTER TABLE clinlims.sample ADD IF NOT EXISTS consent_recorded_at TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE clinlims.sample ADD IF NOT EXISTS consent_recorded_by numeric(10);
ALTER TABLE clinlims.sample ADD CONSTRAINT fk_sample_consent_recorded_by FOREIGN KEY (consent_recorded_by) REFERENCES system_user (id);
CREATE INDEX IF NOT EXISTS idx_sample_consent_provided ON clinlims.sample(consent_provided);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/017-add-informed-consent-fields.xml::OGC-557-2026-04-17-1::herbertYiga
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
