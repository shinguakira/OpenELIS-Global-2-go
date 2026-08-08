-- source: liquibase liquibase/2.1.x.x/fix_migration_issues.xml::202303245::CIV Developer Group
-- +goose Up
-- +goose StatementBegin
-- Update  localisation english label
update clinlims.localization set english = 'Serum' where english
            = 'Sérum';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/fix_migration_issues.xml::202303245::CIV Developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
