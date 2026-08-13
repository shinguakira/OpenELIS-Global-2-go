-- source: liquibase liquibase/3.3.x.x/013-remove-short-code-columns.xml::storage-012-remove-shelf-short-code::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
-- Drop unique constraint if it exists
ALTER TABLE clinlims.storage_shelf
            DROP CONSTRAINT IF EXISTS uk_shelf_parent_short_code;
ALTER TABLE storage_shelf DROP COLUMN IF EXISTS short_code;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/013-remove-short-code-columns.xml::storage-012-remove-shelf-short-code::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
