-- source: liquibase liquibase/3.4.x.x/014-sample-collection-workflow-menus.xml::nav-001-create-order-workflow-parent::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Create Order Workflow parent menu under Sample menu with dashboard as action
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_sample'), 1, 'menu_order_workflow', '/order', 'sidenav.label.addorder', 'sidenav.label.addorder', FALSE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/014-sample-collection-workflow-menus.xml::nav-001-create-order-workflow-parent::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
