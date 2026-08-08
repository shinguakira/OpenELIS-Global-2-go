-- source: liquibase liquibase/2.7.x.x/add_admin_default_roles.xml::1::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- Add default Lab Unit Roles to admin User
INSERT INTO clinlims.user_lab_unit_roles (system_user_id, last_updated) VALUES ((SELECT id FROM system_user WHERE login_name = 'admin'), NOW()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.lab_unit_role_map (lab_unit) VALUES ('AllLabUnits') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.lab_roles (lab_unit_role_map_id, role) VALUES ((SELECT MAX(lab_unit_role_map_id) FROM lab_unit_role_map), (SELECT id FROM system_role WHERE name = 'Reception')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.lab_roles (lab_unit_role_map_id, role) VALUES ((SELECT MAX(lab_unit_role_map_id) FROM lab_unit_role_map), (SELECT id FROM system_role WHERE name = 'Results')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.lab_roles (lab_unit_role_map_id, role) VALUES ((SELECT MAX(lab_unit_role_map_id) FROM lab_unit_role_map), (SELECT id FROM system_role WHERE name = 'Validation')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.lab_roles (lab_unit_role_map_id, role) VALUES ((SELECT MAX(lab_unit_role_map_id) FROM lab_unit_role_map), (SELECT id FROM system_role WHERE name = 'Reports')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.lab_unit_roles (system_user_id, lab_unit_role_map_id) VALUES ((SELECT MAX(system_user_id) FROM user_lab_unit_roles), (SELECT MAX(lab_unit_role_map_id) FROM lab_unit_role_map)) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_admin_default_roles.xml::1::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
