-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::reorganize-storage-menu-2::mherman22
-- +goose Up
-- +goose StatementBegin
-- Add Storage Management Dashboard as child of Storage menu
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_storage'), '1', 'menu_storage_management', '/Storage', 'sidenav.label.storage.management', 'sidenav.label.storage.management.tooltip', FALSE, TRUE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::reorganize-storage-menu-2::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
