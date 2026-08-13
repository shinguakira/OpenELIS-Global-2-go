-- source: liquibase liquibase/qc/007-add-qc-menu-items.xml::qc-017-add-menu-qc-parent::openelisglobal
-- +goose Up
-- +goose StatementBegin
-- Adds Quality Control sub-menu under Analyzers
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_analyzers'), '5', 'menu_analyzers_qc', '/analyzers/qc/db', 'analyzer.navigation.qc', 'analyzer.navigation.qc', FALSE, TRUE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/qc/007-add-qc-menu-items.xml::qc-017-add-menu-qc-parent::openelisglobal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
