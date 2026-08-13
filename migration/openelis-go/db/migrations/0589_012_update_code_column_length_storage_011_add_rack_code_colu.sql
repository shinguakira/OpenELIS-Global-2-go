-- source: liquibase liquibase/3.3.x.x/012-update-code-column-length.xml::storage-011-add-rack-code-column::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
ALTER TABLE storage_rack ADD IF NOT EXISTS code VARCHAR(10);
-- Populate code from short_code if it exists, otherwise generate from label
            UPDATE clinlims.storage_rack
            SET code = COALESCE(
                short_code,
                UPPER(REGEXP_REPLACE(label, '[^A-Z0-9_-]', '', 'g'))
            )
            WHERE code IS NULL;
-- Truncate to 10 chars if needed
            UPDATE clinlims.storage_rack
            SET code = LEFT(code, 10)
            WHERE LENGTH(code) > 10;
ALTER TABLE storage_rack ALTER COLUMN  code SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/012-update-code-column-length.xml::storage-011-add-rack-code-column::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
