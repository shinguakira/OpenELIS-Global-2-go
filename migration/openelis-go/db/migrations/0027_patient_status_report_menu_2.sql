-- source: liquibase liquibase/2.1.x.x/patient_status_report_menu.xml::2::ctsteele
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.menu SET action_url = '/Report.do?type=patient&report=patientCILNSP_vreduit' WHERE element_id='menu_reports_status_patient';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/patient_status_report_menu.xml::2::ctsteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
