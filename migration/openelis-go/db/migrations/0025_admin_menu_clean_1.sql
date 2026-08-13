-- source: liquibase liquibase/2.1.x.x/admin_menu_clean.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
DELETE FROM clinlims.menu WHERE element_id = 'menu_administration_test_management' OR element_id = 'menu_administration_report_management';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/admin_menu_clean.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
