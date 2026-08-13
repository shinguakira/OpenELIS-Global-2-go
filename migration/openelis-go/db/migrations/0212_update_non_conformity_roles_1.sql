-- source: liquibase liquibase/2.6.x.x/update_non_conformity_roles.xml::1::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- update Validation, Results and Reception Roles to have Acces to Non Conformity Tab
INSERT INTO clinlims.system_role_module (id, has_select, has_add, has_update, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), 'Y', 'Y', 'Y', (SELECT id FROM clinlims.system_role WHERE name = 'Validation'), (SELECT id FROM clinlims.system_module WHERE name = 'NonConformity')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/ReportNonConformingEvent', (SELECT id FROM clinlims.system_module WHERE name = 'NonConformity')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/ViewNonConformingEvent', (SELECT id FROM clinlims.system_module WHERE name = 'NonConformity')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/NCECorrectiveAction', (SELECT id FROM clinlims.system_module WHERE name = 'NonConformity')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/update_non_conformity_roles.xml::1::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
