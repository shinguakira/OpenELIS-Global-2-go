-- source: liquibase liquibase/2.6.x.x/add_order_priority.xml::3::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- add Option 'By Priority' to the WorkPlan menu
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_workplan'), '4', 'menu_workplan_priority', '/WorkPlanByPriotiy?type=priority', 'banner.menu.workplan.priority', 'tooltip.bannner.menu.workplan.priority', 'false', 'true') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/add_order_priority.xml::3::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
