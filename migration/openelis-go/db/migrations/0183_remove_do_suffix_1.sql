-- source: liquibase liquibase/2.6.x.x/remove_do_suffix.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- remove all .do suffixes
UPDATE clinlims.menu SET action_url = REPLACE(action_url, '.do', '');
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/remove_do_suffix.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
