-- source: liquibase liquibase/3.5.x.x/026-register-audit-ref-tables-for-p0-services.xml::026-register-audit-ref-tables-3::oe
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.reference_tables (id, name, keep_history, lastupdated) VALUES (nextval('clinlims.reference_tables_seq'), 'qc_control_lot', 'Y', NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/026-register-audit-ref-tables-for-p0-services.xml::026-register-audit-ref-tables-3::oe
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
