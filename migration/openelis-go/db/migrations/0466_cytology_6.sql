-- source: liquibase liquibase/2.8.x.x/cytology.xml::6::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Create Cytopathologist Role and permission for pages
INSERT INTO clinlims.system_role (id, name, description, is_grouping_role, grouping_parent, display_key, active, editable) VALUES (nextval('clinlims.system_role_seq'), 'Cytopathologist', 'Access to Cytology Page', 'false', (SELECT id FROM system_role WHERE name = 'Global Roles'), 'role.Cytopathologist', 'true', 'false') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module (id, name, description) VALUES (nextval('clinlims.system_module_seq'), 'Cytology', 'Cytology pages') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_role_module (id, has_select, has_add, has_update, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), 'Y', 'Y', 'Y', (SELECT id FROM clinlims.system_role WHERE name = 'Cytopathologist'), (SELECT id FROM clinlims.system_module WHERE name = 'Cytology')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/CytologyOrderEntry', (SELECT id FROM clinlims.system_module WHERE name = 'Cytology')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/CytologyDashboard', (SELECT id FROM clinlims.system_module WHERE name = 'Cytology')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/CytologyCaseView', (SELECT id FROM clinlims.system_module WHERE name = 'Cytology')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/cytology.xml::6::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
