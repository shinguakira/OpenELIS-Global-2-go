-- source: liquibase liquibase/2.7.x.x/add_TB_menu.xml::23102505::CIV developer Group
-- +goose Up
-- +goose StatementBegin
-- Add menu entry for tb report
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_tb_report'), '1', 'menu_tb_order_report', '/Report?type=patient&report=TBOrderReport', 'openreports.activity.title', 'openreports.activity.title', 'false', 'true') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_TB_menu.xml::23102505::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
