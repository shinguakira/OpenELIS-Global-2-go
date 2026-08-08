-- source: liquibase liquibase/3.3.x.x/008-add-short-code-columns.xml::storage-010a-make-rack-short-code-required::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
-- Update existing NULL values to unique temporary values based on ID
            -- Users must update these via Edit form to set proper short codes
UPDATE clinlims.storage_rack
            SET short_code = 'TMP' || LPAD(id::text, 7, '0')
            WHERE short_code IS NULL;

ALTER TABLE storage_rack ADD CONSTRAINT uk_rack_parent_short_code UNIQUE (parent_shelf_id, short_code);

ALTER TABLE storage_rack ALTER COLUMN  short_code SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/008-add-short-code-columns.xml::storage-010a-make-rack-short-code-required::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
