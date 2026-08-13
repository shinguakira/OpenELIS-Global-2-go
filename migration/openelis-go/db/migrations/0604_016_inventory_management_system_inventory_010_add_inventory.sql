-- source: liquibase liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-010-add-inventory-management-submenu::mherman22
-- +goose Up
-- +goose StatementBegin
-- Add Inventory Management Dashboard as child of Inventory menu
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_inventory'), '1', 'menu_inventory_management', '/inventory', 'sidenav.label.inventory.management', 'sidenav.label.inventory.management.tooltip', FALSE, TRUE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-010-add-inventory-management-submenu::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
