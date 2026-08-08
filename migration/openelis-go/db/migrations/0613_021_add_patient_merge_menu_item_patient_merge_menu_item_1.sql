-- source: liquibase liquibase/3.3.x.x/021-add-patient-merge-menu-item.xml::patient-merge-menu-item-1::claude
-- +goose Up
-- +goose StatementBegin
-- Adds Patient Merge menu item under Patient menu
INSERT INTO clinlims.menu (id, presentation_order, element_id, action_url, parent_id, display_key, tool_tip_key, new_window, is_active, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), '4', 'menu_patient_merge', '/PatientMerge', (SELECT id FROM clinlims.menu WHERE element_id = 'menu_patient'), 'banner.menu.patient.merge', 'banner.menu.patient.merge.tooltip', FALSE, TRUE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/021-add-patient-merge-menu-item.xml::patient-merge-menu-item-1::claude
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
