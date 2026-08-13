-- source: liquibase liquibase/3.4.x.x/014-sample-collection-workflow-menus.xml::nav-001b-create-order-dashboard-menu::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add Order Dashboard menu entry
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_order_workflow'), 0, 'menu_order_dashboard', '/order', 'sidenav.label.order.dashboard', 'sidenav.label.order.dashboard', FALSE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/014-sample-collection-workflow-menus.xml::nav-001b-create-order-dashboard-menu::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
