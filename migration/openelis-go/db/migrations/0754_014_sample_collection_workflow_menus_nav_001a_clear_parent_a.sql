-- source: liquibase liquibase/3.4.x.x/014-sample-collection-workflow-menus.xml::nav-001a-clear-parent-action-url::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Clear action_url on parent menu so it expands instead of navigates
UPDATE clinlims.menu SET action_url = '' WHERE element_id = 'menu_order_workflow';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/014-sample-collection-workflow-menus.xml::nav-001a-clear-parent-action-url::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
