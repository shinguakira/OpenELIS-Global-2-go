-- source: liquibase liquibase/2.8.x.x/panel_item_fix.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- remove panel items that test no longer exist in db
DELETE FROM clinlims.panel_item WHERE test_id NOT IN (SELECT id FROM clinlims.test);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/panel_item_fix.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
