-- source: liquibase liquibase/3.2.x.x/sample-management.xml::sample-mgmt-009-add-menu-entry::dev-team
-- +goose Up
-- +goose StatementBegin
-- Add Sample Management menu entry as child of Generic Sample with dynamic ID generation
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_generic_sample'), 4, 'menu_sample_management', '/SampleManagement', 'banner.menu.sampleManagement', 'banner.menu.sampleManagement.tooltip', FALSE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/sample-management.xml::sample-mgmt-009-add-menu-entry::dev-team
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
