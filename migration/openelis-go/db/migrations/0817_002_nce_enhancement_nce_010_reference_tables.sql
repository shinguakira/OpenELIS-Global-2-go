-- source: liquibase liquibase/3.5.x.x/002-nce-enhancement.xml::nce-010-reference-tables::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add reference table entries for NCE tables
INSERT INTO clinlims.reference_tables (id, name, keep_history, is_hl7_encoded)
            SELECT COALESCE(MAX(id), 0) + 1, 'NCE_ATTACHMENT', 'Y', 'N'
            FROM clinlims.reference_tables ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/002-nce-enhancement.xml::nce-010-reference-tables::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
