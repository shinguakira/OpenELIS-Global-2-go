-- source: liquibase liquibase/2.3.x.x/minor_fixes.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
DELETE FROM clinlims.site_information WHERE name='non-conformity signature';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/minor_fixes.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
