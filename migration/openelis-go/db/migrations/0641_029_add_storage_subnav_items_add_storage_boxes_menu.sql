-- source: liquibase liquibase/3.3.x.x/029-add-storage-subnav-items.xml::add-storage-boxes-menu::navbar-extraction
-- +goose Up
-- +goose StatementBegin
-- Add Boxes tab under Storage Management
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_storage_management'), '6', 'menu_storage_boxes', '/Storage/boxes', 'storage.nav.boxes', 'storage.nav.boxes.tooltip', FALSE, TRUE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/029-add-storage-subnav-items.xml::add-storage-boxes-menu::navbar-extraction
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
