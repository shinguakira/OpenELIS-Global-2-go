-- source: liquibase liquibase/2.3.x.x/minor_fixes.xml::3::csteele
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.result_limits SET min_age = FLOOR(min_age/12*365) WHERE min_age != 'infinity';

UPDATE clinlims.result_limits SET max_age = FLOOR(max_age/12*365) WHERE max_age != 'infinity';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/minor_fixes.xml::3::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
