-- source: liquibase liquibase/2.8.x.x/add_antimicrobial_resistance_test_column.xml::2::rossumg
-- +goose Up
-- +goose StatementBegin
-- Add menu entry for whonet export
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_reports'), '99', 'menu_reports_whonet_export', '/Report?type=patient&report=ExportWHONETReportByDate', 'reports.export.whonet', 'tooltip.reports.export.whonet', 'false', 'true') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/add_antimicrobial_resistance_test_column.xml::2::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
