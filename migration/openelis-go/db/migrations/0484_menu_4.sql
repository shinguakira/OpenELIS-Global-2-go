-- source: liquibase liquibase/2.8.x.x/menu.xml::4::csteele
-- +goose Up
-- +goose StatementBegin
-- update study tab menu option
UPDATE clinlims.menu SET is_active = (SELECT (si.value = 'true') FROM clinlims.site_information si WHERE si.name = 'Study Management tab') WHERE element_id IN ('menu_sample_create', 'menu_sample_create_initial', 'menu_sample_create_double', 'menu_sample_consult', 'menu_patient_create', 'menu_patient_create_initial', 'menu_patient_create_double', 'menu_patient_edit', 'menu_patient_consult', 'menu_reports_study', 'menu_reports_patients', 'menu_reports_arv', 'menu_reports_arv_initial1', 'menu_reports_arv_initial2', 'menu_reports_arv_followup1', 'menu_reports_arv_followup2', 'menu_reports_eid', 'menu_reports_eid_version1', 'menu_reports_eid_version2', 'menu_reports_indeterminate', 'menu_reports_indeterminate_version1', 'menu_reports_indeterminate_version2', 'menu_reports_indeterminate_location', 'menu_reports_special', 'menu_reports_patient_collection', 'menu_reports_patient_associated', 'menu_reports_indicator', 'menu_reports_indicator_performance', 'menu_reports_validation_backlog.study', 'menu_reports_nonconformity.study', 'menu_reports_nonconformity_date.study', 'menu_reports_nonconformity_section.study', 'menu_reports_nonconformity_notification.study', 'menu_reports_followupRequired_ByLocation.study', 'menu_reports_export', 'menu_reports_auditTrail.study', 'menu_reports_arv_all', 'menu_reports_vl', 'menu_reports_vl_version1', 'menu_reports_nonconformity.Labno', 'menu_resultvalidation_study', 'menu_resultvalidation_immunology', 'menu_resultvalidation_biochemistry', 'menu_resultvalidation_serology', 'menu_resultvalidation_dnapcr', 'menu_resultvalidation_virology', 'menu_resultvalidation_viralload', 'menu_resultvalidation_genotyping');

DELETE FROM clinlims.site_information WHERE name='Study Management tab';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/menu.xml::4::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
