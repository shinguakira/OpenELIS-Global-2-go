-- source: liquibase liquibase/3.3.x.x/029-add-storage-subnav-items.xml::update-storage-management-no-action::navbar-extraction
-- +goose Up
-- +goose StatementBegin
-- Remove action_url from Storage Management to make it expandable with child items
UPDATE clinlims.menu SET action_url = '' WHERE element_id = 'menu_storage_management';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/029-add-storage-subnav-items.xml::update-storage-management-no-action::navbar-extraction
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
