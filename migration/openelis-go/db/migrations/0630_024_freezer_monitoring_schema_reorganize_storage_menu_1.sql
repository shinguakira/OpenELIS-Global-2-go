-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::reorganize-storage-menu-1::mherman22
-- +goose Up
-- +goose StatementBegin
-- Remove action_url from Storage menu to make it expandable with child items
UPDATE clinlims.menu SET action_url = '' WHERE element_id = 'menu_storage';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::reorganize-storage-menu-1::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
