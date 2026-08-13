-- source: liquibase liquibase/3.2.x.x/notebook.xml::create-notebook-reference-table-entry::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Creating NOTEBOOK reference table entry to enable audit trail tracking.
INSERT INTO clinlims.reference_tables (id, name, keep_history, is_hl7_encoded, lastupdated) VALUES (nextval('clinlims.reference_tables_seq'), 'NOTEBOOK', 'Y', 'N', NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/notebook.xml::create-notebook-reference-table-entry::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
