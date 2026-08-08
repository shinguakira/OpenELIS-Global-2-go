-- source: liquibase liquibase/2.1.x.x/sample_batch_entry_module_permission.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/SamplePatientEntryBatch', (SELECT id FROM clinlims.system_module WHERE name = 'SampleBatchEntry')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/sample_batch_entry_module_permission.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
