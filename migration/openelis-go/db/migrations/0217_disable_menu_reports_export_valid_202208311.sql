-- source: liquibase liquibase/2.6.x.x/disable_menu_reports_export_valid.xml::202208311::CIV developer Group
-- +goose Up
-- +goose StatementBegin
-- Disable menu option for export study valid result
UPDATE clinlims.menu SET is_active = 'false' WHERE element_id = 'menu_reports_export_valid';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/disable_menu_reports_export_valid.xml::202208311::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
