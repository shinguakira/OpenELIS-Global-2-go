-- source: liquibase liquibase/3.2.x.x/sample_item.xml::add-voided-22-10-2025::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add voided column to sample_item table
ALTER TABLE sample_item ADD IF NOT EXISTS voided BOOLEAN DEFAULT FALSE;
UPDATE sample_item SET voided = FALSE;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/sample_item.xml::add-voided-22-10-2025::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
