-- source: liquibase liquibase/2.6.x.x/update_roles.xml::2::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- update Recetion Role
UPDATE clinlims.system_role SET grouping_parent = (SELECT id FROM system_role WHERE name = 'Lab Unit Roles'), name = 'Reception' WHERE id=(SELECT id FROM system_role WHERE name = 'Intake');
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/update_roles.xml::2::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
