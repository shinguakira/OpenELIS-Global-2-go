-- source: liquibase liquibase/2.8.x.x/update_default_setting.xml::3::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Create AdminReportPrint Module
INSERT INTO clinlims.system_module (id, name, description) VALUES (nextval('clinlims.system_module_seq'), 'Report:RoutineExport', 'Report => Routine => Eport Routine') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_role_module (id, has_select, has_add, has_update, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), 'Y', 'Y', 'Y', (SELECT id FROM clinlims.system_role WHERE name = 'Global Administrator'), (SELECT id FROM clinlims.system_module WHERE name = 'Report:RoutineExport')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_param (id, name, value) VALUES (nextval('clinlims.system_module_param_seq'), 'report', 'CISampleRoutineExport') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id, system_module_param_id) VALUES (nextval('clinlims.system_module_url_seq'), '/ReportPrint', (SELECT id FROM clinlims.system_module WHERE name = 'Report:RoutineExport'), (SELECT id FROM system_module_param WHERE value = 'CISampleRoutineExport')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id, system_module_param_id) VALUES (nextval('clinlims.system_module_url_seq'), '/Report', (SELECT id FROM clinlims.system_module WHERE name = 'Report:RoutineExport'), (SELECT id FROM system_module_param WHERE value = 'CISampleRoutineExport')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/update_default_setting.xml::3::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
