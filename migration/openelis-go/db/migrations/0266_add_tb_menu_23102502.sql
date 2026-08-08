-- source: liquibase liquibase/2.7.x.x/add_TB_menu.xml::23102502::CIV developer Group
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.menu SET action_url = '' WHERE id=(SELECT id FROM clinlims.menu WHERE element_id = 'menu_tb_report' limit 1);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_TB_menu.xml::23102502::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
