-- source: liquibase liquibase/2.8.x.x/update_default_setting.xml::1::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- Add default Program Global Roles to admin User
INSERT INTO clinlims.system_user_role (system_user_id, role_id) VALUES ((SELECT id FROM system_user WHERE login_name = 'admin'), (SELECT id FROM system_role WHERE name = 'Cytopathologist')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_user_role (system_user_id, role_id) VALUES ((SELECT id FROM system_user WHERE login_name = 'admin'), (SELECT id FROM system_role WHERE name = 'Pathologist')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/update_default_setting.xml::1::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
