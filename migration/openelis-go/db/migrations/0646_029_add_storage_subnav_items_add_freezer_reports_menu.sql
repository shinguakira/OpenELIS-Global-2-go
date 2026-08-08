-- source: liquibase liquibase/3.3.x.x/029-add-storage-subnav-items.xml::add-freezer-reports-menu::navbar-extraction
-- +goose Up
-- +goose StatementBegin
-- Add Reports tab under Cold Storage Monitoring
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_freezer_monitoring'), '4', 'menu_freezer_reports', '/FreezerMonitoring?tab=3', 'freezer.nav.reports', 'freezer.nav.reports.tooltip', FALSE, TRUE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/029-add-storage-subnav-items.xml::add-freezer-reports-menu::navbar-extraction
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
