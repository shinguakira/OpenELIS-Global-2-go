-- source: liquibase liquibase/3.5.x.x/019-update-consent-recorded-by-to-varchar.xml::OGC-558-2026-04-30-2::herbertYiga
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.sample ALTER COLUMN consent_recorded_by TYPE VARCHAR(255) USING (consent_recorded_by::VARCHAR(255));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/019-update-consent-recorded-by-to-varchar.xml::OGC-558-2026-04-30-2::herbertYiga
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
