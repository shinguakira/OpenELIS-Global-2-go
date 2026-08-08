-- source: liquibase liquibase/2.0.x.x/patient_contact.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.reference_tables (id, name, keep_history, is_hl7_encoded) VALUES (nextval('clinlims.reference_tables_seq'), 'PATIENT_CONTACT', 'Y', 'N') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.0.x.x/patient_contact.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
