-- source: liquibase liquibase/2.3.x.x/results_range.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.system_module (id, name, description) VALUES (nextval('clinlims.system_module_seq'), 'RangeResults', 'view results by accession range') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_role_module (id, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), (SELECT id FROM clinlims.system_role WHERE name = 'Results Admin'), (SELECT id FROM clinlims.system_module WHERE name = 'RangeResults')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_role_module (id, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), (SELECT id FROM clinlims.system_role WHERE name = 'Results entry'), (SELECT id FROM clinlims.system_module WHERE name = 'RangeResults')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_role_module (id, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), (SELECT id FROM clinlims.system_role WHERE name = 'Results modifier'), (SELECT id FROM clinlims.system_module WHERE name = 'RangeResults')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/RangeResults', (SELECT id FROM clinlims.system_module WHERE name = 'RangeResults')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/results_range.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
