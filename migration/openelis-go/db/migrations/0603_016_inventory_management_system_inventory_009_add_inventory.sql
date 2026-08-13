-- source: liquibase liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-009-add-inventory-menu-item::mherman22
-- +goose Up
-- +goose StatementBegin
-- Add top-level Inventory menu entry (expandable parent, no action URL)
INSERT INTO clinlims.menu (id, presentation_order, element_id, display_key, tool_tip_key, new_window, is_active, action_url, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), '127', 'menu_inventory', 'sidenav.label.inventory', 'sidenav.label.inventory.tooltip', FALSE, TRUE, '', TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-009-add-inventory-menu-item::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
