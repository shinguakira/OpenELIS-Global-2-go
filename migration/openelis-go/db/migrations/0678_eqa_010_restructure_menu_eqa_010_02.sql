-- source: liquibase liquibase/3.3.x.x/eqa-010-restructure-menu.xml::eqa-010-02::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Deactivate old EQA child menu items
UPDATE clinlims.menu SET is_active = FALSE WHERE parent_id = (SELECT id FROM clinlims.menu WHERE element_id = 'menu_eqa');
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-010-restructure-menu.xml::eqa-010-02::eqa-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
