-- source: liquibase liquibase/2.8.x.x/sample_program_audit_trail.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- Create program sample reference tables entries
INSERT INTO clinlims.reference_tables (id, name, keep_history, is_hl7_encoded, lastupdated) VALUES (nextval('clinlims.reference_tables_seq'), 'PATHOLOGY_SAMPLE', 'Y', 'N', NOW()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.reference_tables (id, name, keep_history, is_hl7_encoded, lastupdated) VALUES (nextval('clinlims.reference_tables_seq'), 'IMMUNOHISTOCHEMISTRY_SAMPLE', 'Y', 'N', NOW()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.reference_tables (id, name, keep_history, is_hl7_encoded, lastupdated) VALUES (nextval('clinlims.reference_tables_seq'), 'CYTOLOGY_SAMPLE', 'Y', 'N', NOW()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.reference_tables (id, name, keep_history, is_hl7_encoded, lastupdated) VALUES (nextval('clinlims.reference_tables_seq'), 'PROGRAM_SAMPLE', 'Y', 'N', NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/sample_program_audit_trail.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
