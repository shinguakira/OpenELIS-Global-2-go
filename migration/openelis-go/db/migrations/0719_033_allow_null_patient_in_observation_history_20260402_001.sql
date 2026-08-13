-- source: liquibase liquibase/3.3.x.x/033-allow-null-patient-in-observation-history.xml::20260402-001::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Allow null patient_id in observation_history for environmental workflow samples
ALTER TABLE clinlims.observation_history ALTER COLUMN  patient_id DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/033-allow-null-patient-in-observation-history.xml::20260402-001::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
