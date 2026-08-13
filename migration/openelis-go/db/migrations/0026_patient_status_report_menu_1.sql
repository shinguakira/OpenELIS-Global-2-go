-- source: liquibase liquibase/2.1.x.x/patient_status_report_menu.xml::1::rossumg
-- +goose Up
-- +goose StatementBegin
DELETE FROM clinlims.menu WHERE element_id in ('menu_reports_status_patient.vreduit','menu_reports_status_patient.classique');
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/patient_status_report_menu.xml::1::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
