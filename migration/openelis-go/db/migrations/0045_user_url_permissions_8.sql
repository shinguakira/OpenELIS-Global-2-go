-- source: liquibase liquibase/2.3.x.x/user_url_permissions.xml::8::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.system_module (id, name, description) VALUES (nextval('clinlims.system_module_seq'), 'ReportCovid', 'covid report pages') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_role_module (id, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), (SELECT id FROM clinlims.system_role WHERE name = 'Maintenance Admin'), (SELECT id FROM clinlims.system_module WHERE name = 'ReportCovid')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_role_module (id, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), (SELECT id FROM clinlims.system_role WHERE name = 'Reports'), (SELECT id FROM clinlims.system_module WHERE name = 'ReportCovid')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_param (id, name, value) VALUES (nextval('clinlims.system_module_param_seq'), 'report', 'covidResultsReport') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id, system_module_param_id) VALUES (nextval('clinlims.system_module_url_seq'), '/ReportPrint', (SELECT id FROM clinlims.system_module WHERE name = 'ReportCovid'), (SELECT id FROM system_module_param WHERE value = 'covidResultsReport')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id, system_module_param_id) VALUES (nextval('clinlims.system_module_url_seq'), '/Report', (SELECT id FROM clinlims.system_module WHERE name = 'ReportCovid'), (SELECT id FROM system_module_param WHERE value = 'covidResultsReport')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/user_url_permissions.xml::8::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
