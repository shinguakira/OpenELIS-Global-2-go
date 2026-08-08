-- source: liquibase liquibase/2.7.x.x/add_TB_menu.xml::2023092401::CIV developer Group
-- +goose Up
-- +goose StatementBegin
-- Add menu entry for microbiology report
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_reports'), '3', 'menu_report_microbiology', 'banner.menu.microbiology', 'tooltip.bannner.menu.microbiology', 'false', 'true') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_TB_menu.xml::2023092401::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
