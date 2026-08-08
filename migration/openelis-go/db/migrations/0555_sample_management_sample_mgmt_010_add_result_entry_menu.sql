-- source: liquibase liquibase/3.2.x.x/sample-management.xml::sample-mgmt-010-add-result-entry-menu::dev-team
-- +goose Up
-- +goose StatementBegin
-- Add Result Entry menu entry as child of Results menu
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_results'), 1, 'menu_sample_management', '/ResultEntry', 'result.entry.label', 'result.entry.label', FALSE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/sample-management.xml::sample-mgmt-010-add-result-entry-menu::dev-team
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
