-- source: liquibase liquibase/2.8.x.x/update_default_setting.xml::4::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- update action url of Routine Export report
UPDATE clinlims.menu SET action_url = '/Report?type=routine&report=CISampleRoutineExport' WHERE element_id='menu_reports_export_routine';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/update_default_setting.xml::4::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
