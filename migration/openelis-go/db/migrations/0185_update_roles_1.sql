-- source: liquibase liquibase/2.6.x.x/update_roles.xml::1::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- update existing roles and Create new ones
INSERT INTO clinlims.system_role (id, name, description, is_grouping_role, display_key, active, editable) VALUES (nextval('clinlims.system_role_seq'), 'Global Roles', 'Grouping Role for all Global Roles', 'true', 'global', 'true', 'false') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_role (id, name, description, is_grouping_role, display_key, active, editable) VALUES (nextval('clinlims.system_role_seq'), 'Lab Unit Roles', 'Grouping Role for all Lab Unit Roles', 'true', 'lab.unit', 'true', 'false') ON CONFLICT DO NOTHING;
UPDATE clinlims.system_role SET grouping_parent = (SELECT id FROM system_role WHERE name = 'Global Roles'), name = 'Global Administrator' WHERE id=1;
UPDATE clinlims.system_role SET grouping_parent = (SELECT id FROM system_role WHERE name = 'Global Roles'), name = 'User Account Administrator' WHERE id=2;
UPDATE clinlims.system_role SET grouping_parent = (SELECT id FROM system_role WHERE name = 'Global Roles') WHERE id=(SELECT id FROM system_role WHERE name = 'Audit Trail');
UPDATE clinlims.system_role SET grouping_parent = (SELECT id FROM system_role WHERE name = 'Lab Unit Roles'), name = 'Reception' WHERE id=(SELECT id FROM system_role WHERE name = 'Reception');
UPDATE clinlims.system_role SET grouping_parent = (SELECT id FROM system_role WHERE name = 'Lab Unit Roles'), name = 'Results' WHERE id=(SELECT id FROM system_role WHERE name = 'Results' or name = 'Results entry' limit 1);
UPDATE clinlims.system_role SET grouping_parent = (SELECT id FROM system_role WHERE name = 'Lab Unit Roles'), name = 'Validation' WHERE id=(SELECT id FROM system_role WHERE name = 'Validation' or name='Validator' limit 1);
UPDATE clinlims.system_role SET grouping_parent = (SELECT id FROM system_role WHERE name = 'Lab Unit Roles') WHERE id=(SELECT id FROM system_role WHERE name = 'Reports');
UPDATE clinlims.system_role SET active = 'false' WHERE id=(SELECT id FROM system_role WHERE name = 'Inventory mgr');
UPDATE clinlims.system_role SET active = 'false' WHERE id=(SELECT id FROM system_role WHERE name = 'Results modifier');
UPDATE clinlims.system_role SET active = 'false' WHERE id=(SELECT id FROM system_role WHERE name = 'Test Management');
UPDATE clinlims.system_role SET active = 'false' WHERE id=(SELECT id FROM system_role WHERE name = 'Results Admin');
UPDATE clinlims.system_role SET active = 'false' WHERE id=(SELECT id FROM system_role WHERE name = 'Identifying Information Edit');
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/update_roles.xml::1::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
