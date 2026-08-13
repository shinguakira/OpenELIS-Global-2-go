-- source: liquibase liquibase/2.1.x.x/fix_migration_issues.xml::202303244::CIV Developer Group
-- +goose Up
-- +goose StatementBegin
-- Update  next value for observation_history_type_seq after database migration
SELECT setval('clinlims.observation_history_type_seq', CAST
            ((SELECT coalesce(MAX(id),0) FROM clinlims.observation_history_type)
            AS BIGINT) + 1);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/fix_migration_issues.xml::202303244::CIV Developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
