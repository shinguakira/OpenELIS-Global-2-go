-- source: liquibase liquibase/2.6.x.x/add_menu_study_electronic_order.xml::202208291::CIV developer Group
-- +goose Up
-- +goose StatementBegin
-- Add menu entry for study electronic orders
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_sample_create'), '4', 'menu_study_sample_eorder', '/StudyElectronicOrders', 'banner.menu.study.eorders', 'tooltip.bannner.menu.study.eorders', 'false', 'true') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/add_menu_study_electronic_order.xml::202208291::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
