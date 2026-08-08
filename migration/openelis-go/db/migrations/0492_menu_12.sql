-- source: liquibase liquibase/2.8.x.x/menu.xml::12::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- update results search menu options
DELETE FROM clinlims.menu WHERE element_id='menu_results_search_testdate';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/menu.xml::12::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
