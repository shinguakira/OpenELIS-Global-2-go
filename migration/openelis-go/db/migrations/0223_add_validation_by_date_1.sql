-- source: liquibase liquibase/2.7.x.x/add_validation_by_date.xml::1::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- Add Validation Search By Date to the Validation Tab menu
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_resultvalidation'), '22', 'menu_resultvalidation_date', '/ResultValidationByTestDate', 'menu.validation.date', 'tooltip.validation.date', 'false', 'true') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_validation_by_date.xml::1::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
