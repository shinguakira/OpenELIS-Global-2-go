-- source: liquibase liquibase/2.8.x.x/menu.xml::13::csteele
-- +goose Up
-- +goose StatementBegin
-- add column to track if Menu shows up in old UI
ALTER TABLE clinlims.menu ADD IF NOT EXISTS hide_in_old_ui BOOLEAN DEFAULT FALSE;
UPDATE clinlims.menu SET hide_in_old_ui = TRUE WHERE element_id in ('menu_billing', 'menu_patienthistory', 'menu_pathology', 'menu_pathologydashboard', 'menu_immunochem', 'menu_immunochemdashboard', 'menu_cytology', 'menu_cytologydashboard', 'menu_results_search_testdate');
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/menu.xml::13::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
