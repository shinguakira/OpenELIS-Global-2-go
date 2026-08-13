-- source: liquibase liquibase/2.1.x.x/user_manual_translation.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.menu SET action_url = '/docs/UserManual' WHERE element_id = 'menu_help_user_manual';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/user_manual_translation.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
