-- source: liquibase liquibase/2.4.x.x/remove_roles.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- remove 'automatic' roles
DELETE FROM clinlims.system_role WHERE name LIKE '%automatic%';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.4.x.x/remove_roles.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
