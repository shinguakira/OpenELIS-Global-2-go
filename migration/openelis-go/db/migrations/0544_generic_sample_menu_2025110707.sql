-- source: liquibase liquibase/3.2.x.x/generic_sample_menu.xml::2025110707::Generic Sample Developer
-- +goose Up
-- +goose StatementBegin
-- Assign Generic Sample module to Reception role
INSERT INTO clinlims.system_role_module (id, has_select, has_add, has_update, has_delete, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), 'Y', 'Y', 'Y', 'N', (SELECT id FROM clinlims.system_role WHERE name = 'Reception'), (SELECT id FROM clinlims.system_module WHERE name = 'GenericSampleView')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/generic_sample_menu.xml::2025110707::Generic Sample Developer
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
