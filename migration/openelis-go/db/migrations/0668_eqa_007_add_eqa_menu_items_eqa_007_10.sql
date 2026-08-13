-- source: liquibase liquibase/3.3.x.x/eqa-007-add-eqa-menu-items.xml::eqa-007-10::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Assign EQAView module to Global Administrator role
INSERT INTO clinlims.system_role_module (id, has_select, has_add, has_update, has_delete, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), 'Y', 'Y', 'Y', 'N', (SELECT id FROM clinlims.system_role WHERE name = 'Global Administrator'), (SELECT id FROM clinlims.system_module WHERE name = 'EQAView')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-007-add-eqa-menu-items.xml::eqa-007-10::eqa-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
