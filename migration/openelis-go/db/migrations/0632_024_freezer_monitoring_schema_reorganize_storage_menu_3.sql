-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::reorganize-storage-menu-3::mherman22
-- +goose Up
-- +goose StatementBegin
-- Move Cold Storage menu as child of Storage menu and update display key
UPDATE clinlims.menu SET display_key = 'sidenav.label.storage.coldstorage', parent_id = (SELECT id FROM clinlims.menu WHERE element_id = 'menu_storage'), presentation_order = '2', tool_tip_key = 'sidenav.label.storage.coldstorage.tooltip' WHERE element_id = 'menu_freezer_monitoring';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::reorganize-storage-menu-3::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
