-- source: liquibase liquibase/2.7.x.x/rename_test_for_csv_export.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- rename Murex/Determine to not collide with other Murex test in CSV export
UPDATE clinlims.localization SET english = 'Determine' WHERE french = 'Determine' AND english = 'Murex';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/rename_test_for_csv_export.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
