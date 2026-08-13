-- source: liquibase liquibase/2.7.x.x/add_admin_default_roles.xml::2::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- Add default Global Roles to admin User
INSERT INTO clinlims.system_user_role (system_user_id, role_id) VALUES ((SELECT id FROM system_user WHERE login_name = 'admin'), (SELECT id FROM system_role WHERE name = 'Analyser Import')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_user_role (system_user_id, role_id) VALUES ((SELECT id FROM system_user WHERE login_name = 'admin'), (SELECT id FROM system_role WHERE name = 'Audit Trail')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_user_role (system_user_id, role_id) VALUES ((SELECT id FROM system_user WHERE login_name = 'admin'), (SELECT id FROM system_role WHERE name = 'Global Administrator')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_user_role (system_user_id, role_id) VALUES ((SELECT id FROM system_user WHERE login_name = 'admin'), (SELECT id FROM system_role WHERE name = 'User Account Administrator')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_user_role (system_user_id, role_id) VALUES ((SELECT id FROM system_user WHERE login_name = 'admin'), (SELECT id FROM system_role WHERE name = 'Reception')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_user_role (system_user_id, role_id) VALUES ((SELECT id FROM system_user WHERE login_name = 'admin'), (SELECT id FROM system_role WHERE name = 'Results')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_user_role (system_user_id, role_id) VALUES ((SELECT id FROM system_user WHERE login_name = 'admin'), (SELECT id FROM system_role WHERE name = 'Validation')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_user_role (system_user_id, role_id) VALUES ((SELECT id FROM system_user WHERE login_name = 'admin'), (SELECT id FROM system_role WHERE name = 'Reports')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_admin_default_roles.xml::2::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
