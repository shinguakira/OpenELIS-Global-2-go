-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::add-freezer-monitoring-menu-item::mherman22
-- +goose Up
-- +goose StatementBegin
-- Add Cold Storage Monitoring entry to the application side navigation
INSERT INTO clinlims.menu (id, presentation_order, element_id, display_key, tool_tip_key, new_window, is_active, action_url) VALUES (nextval('clinlims.menu_seq'), '126', 'menu_freezer_monitoring', 'sidenav.label.coldstorage', 'sidenav.label.coldstorage.tooltip', FALSE, TRUE, '/FreezerMonitoring') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::add-freezer-monitoring-menu-item::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
