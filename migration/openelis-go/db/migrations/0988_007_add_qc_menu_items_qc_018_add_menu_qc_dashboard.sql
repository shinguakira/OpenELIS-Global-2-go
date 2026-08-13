-- source: liquibase liquibase/qc/007-add-qc-menu-items.xml::qc-018-add-menu-qc-dashboard::openelisglobal
-- +goose Up
-- +goose StatementBegin
-- Adds QC Dashboard sub-menu item
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_analyzers_qc'), '1', 'menu_analyzers_qc_dashboard', '/analyzers/qc/db', 'analyzer.navigation.qcDashboard', 'analyzer.navigation.qcDashboard', FALSE, TRUE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/qc/007-add-qc-menu-items.xml::qc-018-add-menu-qc-dashboard::openelisglobal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
