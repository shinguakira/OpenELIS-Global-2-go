-- source: liquibase liquibase/3.3.x.x/eqa-010-restructure-menu.xml::eqa-010-10::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Create Results and Analysis sub-item under EQA Management
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_eqa_mgmt'), '4', 'menu_eqa_mgmt_results', '/EQAResults', 'banner.menu.eqa.mgmt.results', 'banner.menu.eqa.mgmt.results.tooltip', FALSE, TRUE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-010-restructure-menu.xml::eqa-010-10::eqa-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
