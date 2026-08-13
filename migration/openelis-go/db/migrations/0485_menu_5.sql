-- source: liquibase liquibase/2.8.x.x/menu.xml::5::csteele
-- +goose Up
-- +goose StatementBegin
-- delete unused menu option
DELETE FROM clinlims.site_information WHERE name='Data Submission';

DELETE FROM clinlims.site_information WHERE name='Data Sub URL';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/menu.xml::5::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
