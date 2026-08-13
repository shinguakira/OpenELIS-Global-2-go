-- source: liquibase liquibase/2.7.x.x/add_TB_menu.xml::23110301::CIV developer Group
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.menu SET is_active = FALSE WHERE id=(SELECT id FROM clinlims.menu WHERE element_id = 'menu_tb_order_report' limit 1);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_TB_menu.xml::23110301::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
