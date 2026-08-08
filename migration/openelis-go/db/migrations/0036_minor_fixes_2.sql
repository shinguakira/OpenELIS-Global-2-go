-- source: liquibase liquibase/2.3.x.x/minor_fixes.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.login_user SET last_updated = NOW() WHERE last_updated is null;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/minor_fixes.xml::2::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
