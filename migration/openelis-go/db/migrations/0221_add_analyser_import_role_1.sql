-- source: liquibase liquibase/2.7.x.x/add_analyser_import_role.xml::1::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- Create Analyser Import Role
INSERT INTO clinlims.system_role (id, name, description, is_grouping_role, grouping_parent, display_key, active, editable) VALUES (nextval('clinlims.system_role_seq'), 'Analyser Import', 'Acces to Analyser Results Page', 'false', (SELECT id FROM system_role WHERE name = 'Global Roles'), 'role.analyser', 'true', 'false') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_role_module (id, has_select, has_add, has_update, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), 'Y', 'Y', 'Y', (SELECT id FROM clinlims.system_role WHERE name = 'Analyser Import'), (SELECT id FROM clinlims.system_module WHERE name = 'AnalyzerResults')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/AnalyzerResults', (SELECT id FROM clinlims.system_module WHERE name = 'AnalyzerResults')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_analyser_import_role.xml::1::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
