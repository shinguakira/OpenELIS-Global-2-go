-- source: liquibase liquibase/3.3.x.x/eqa-010-restructure-menu.xml::eqa-010-12::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Register new page URLs with EQAView system module
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/EQAOrders', (SELECT id FROM clinlims.system_module WHERE name = 'EQAView')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/EQAMyPrograms', (SELECT id FROM clinlims.system_module WHERE name = 'EQAView')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/EQAParticipants', (SELECT id FROM clinlims.system_module WHERE name = 'EQAView')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/EQAResults', (SELECT id FROM clinlims.system_module WHERE name = 'EQAView')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-010-restructure-menu.xml::eqa-010-12::eqa-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
