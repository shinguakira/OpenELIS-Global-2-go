-- source: liquibase liquibase/3.2.x.x/generic_sample_menu.xml::2025110706::Generic Sample Developer
-- +goose Up
-- +goose StatementBegin
-- Add URLs for Generic Sample module
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/GenericSampleOrder', (SELECT id FROM clinlims.system_module WHERE name = 'GenericSampleView')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/GenericSampleOrderEdit', (SELECT id FROM clinlims.system_module WHERE name = 'GenericSampleView')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/GenericSampleOrderImport', (SELECT id FROM clinlims.system_module WHERE name = 'GenericSampleView')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/rest/GenericSampleOrder', (SELECT id FROM clinlims.system_module WHERE name = 'GenericSampleView')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/rest/GenericSampleOrder/validate', (SELECT id FROM clinlims.system_module WHERE name = 'GenericSampleView')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/rest/GenericSampleOrder/import', (SELECT id FROM clinlims.system_module WHERE name = 'GenericSampleView')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/generic_sample_menu.xml::2025110706::Generic Sample Developer
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
