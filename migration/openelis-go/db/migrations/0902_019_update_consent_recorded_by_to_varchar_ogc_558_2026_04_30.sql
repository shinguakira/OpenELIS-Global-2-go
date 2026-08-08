-- source: liquibase liquibase/3.5.x.x/019-update-consent-recorded-by-to-varchar.xml::OGC-558-2026-04-30-1::herbertYiga
-- +goose Up
-- +goose StatementBegin
ALTER TABLE sample DROP CONSTRAINT fk_sample_consent_recorded_by;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/019-update-consent-recorded-by-to-varchar.xml::OGC-558-2026-04-30-1::herbertYiga
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
