-- source: liquibase liquibase/2.8.x.x/menu.xml::3::csteele
-- +goose Up
-- +goose StatementBegin
-- update patient tab menu option
INSERT INTO clinlims.menu (id, presentation_order, element_id, action_url, parent_id, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), '2', 'menu_patienthistory', '/PatientHistory', (SELECT id FROM clinlims.menu WHERE element_id = 'menu_patient'), 'banner.menu.patienthistory', 'banner.menu.patienthistory.tooltip', FALSE, TRUE) ON CONFLICT DO NOTHING;
UPDATE clinlims.menu SET presentation_order = '3' WHERE element_id IN ('menu_patient_create');
UPDATE clinlims.menu SET is_active = (SELECT (si.value = 'true') FROM clinlims.site_information si WHERE si.name = 'Patient management tab') WHERE element_id IN ('menu_patient', 'menu_patient_add_or_edit', 'menu_patient_study', 'menu_patient_create', 'menu_patient_create_initial', 'menu_patient_create_double', 'menu_patient_edit', 'menu_patient_consult');
DELETE FROM clinlims.site_information WHERE name='Patient management tab';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/menu.xml::3::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
