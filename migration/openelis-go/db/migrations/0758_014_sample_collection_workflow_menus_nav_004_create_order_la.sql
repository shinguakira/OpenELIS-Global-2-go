-- source: liquibase liquibase/3.4.x.x/014-sample-collection-workflow-menus.xml::nav-004-create-order-label-menu::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add Label and Store menu entry
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_order_workflow'), 3, 'menu_order_label', '/order/label', 'sidenav.label.order.label', 'sidenav.label.order.label', FALSE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/014-sample-collection-workflow-menus.xml::nav-004-create-order-label-menu::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
