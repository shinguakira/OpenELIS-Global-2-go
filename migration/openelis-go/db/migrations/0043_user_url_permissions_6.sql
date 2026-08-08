-- source: liquibase liquibase/2.3.x.x/user_url_permissions.xml::6::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.system_module (id, name, description) VALUES (nextval('clinlims.system_module_seq'), 'ValidationConfig', 'ValidationConfig pages') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_role_module (id, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), (SELECT id FROM clinlims.system_role WHERE name = 'Maintenance Admin'), (SELECT id FROM clinlims.system_module WHERE name = 'ValidationConfig')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/ValidationConfigurationMenu', (SELECT id FROM clinlims.system_module WHERE name = 'ValidationConfig')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/ValidationConfiguration', (SELECT id FROM clinlims.system_module WHERE name = 'ValidationConfig')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/user_url_permissions.xml::6::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
