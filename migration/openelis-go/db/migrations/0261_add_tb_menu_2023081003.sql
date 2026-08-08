-- source: liquibase liquibase/2.7.x.x/add_TB_menu.xml::2023081003::CIV developer Group
-- +goose Up
-- +goose StatementBegin
-- Add roles for Microbiology
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/MicrobiologyTb', (SELECT id FROM clinlims.system_module WHERE name = 'MicrobiologyTBView')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_role_module (id, has_select, has_add, has_update, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), 'Y', 'Y', 'Y', (SELECT id FROM clinlims.system_role WHERE name = 'Reception'), (SELECT id FROM clinlims.system_module WHERE name = 'MicrobiologyTBView')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_TB_menu.xml::2023081003::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
