-- source: liquibase liquibase/2.8.x.x/menu.xml::11::csteele
-- +goose Up
-- +goose StatementBegin
-- update action url of report urls
UPDATE clinlims.menu SET action_url = '/RoutineReports' WHERE element_id='menu_reports_routine';

UPDATE clinlims.menu SET action_url = '/StudyReports' WHERE element_id='menu_reports_study';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/menu.xml::11::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
