-- source: liquibase liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-010-verify-data-migrated::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Verify data has been migrated to localization_value table before dropping columns
SELECT 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-010-verify-data-migrated::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
