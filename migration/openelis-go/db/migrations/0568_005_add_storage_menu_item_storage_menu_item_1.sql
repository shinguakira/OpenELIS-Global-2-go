-- source: liquibase liquibase/3.3.x.x/005-add-storage-menu-item.xml::storage-menu-item-1::pmanko
-- +goose Up
-- +goose StatementBegin
-- Adds Storage menu item
INSERT INTO clinlims.menu (id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), '25', 'menu_storage', '/Storage', 'banner.menu.storage', 'banner.menu.storage.tooltip', FALSE, TRUE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/005-add-storage-menu-item.xml::storage-menu-item-1::pmanko
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
