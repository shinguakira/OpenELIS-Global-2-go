-- source: liquibase liquibase/2.8.x.x/pathology.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- Create Pathologist Role and permission for pages
INSERT INTO clinlims.system_role (id, name, description, is_grouping_role, grouping_parent, display_key, active, editable) VALUES (nextval('clinlims.system_role_seq'), 'Pathologist', 'Access to Pathology Page', 'false', (SELECT id FROM system_role WHERE name = 'Global Roles'), 'role.pathologist', 'true', 'false') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module (id, name, description) VALUES (nextval('clinlims.system_module_seq'), 'Pathology', 'pathology pages') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_role_module (id, has_select, has_add, has_update, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), 'Y', 'Y', 'Y', (SELECT id FROM clinlims.system_role WHERE name = 'Pathologist'), (SELECT id FROM clinlims.system_module WHERE name = 'Pathology')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/PathologyOrderEntry', (SELECT id FROM clinlims.system_module WHERE name = 'Pathology')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/PathologyDashboard', (SELECT id FROM clinlims.system_module WHERE name = 'Pathology')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/PathologyCaseView', (SELECT id FROM clinlims.system_module WHERE name = 'Pathology')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/pathology.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
