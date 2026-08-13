-- source: liquibase liquibase/3.3.x.x/011-make-short-code-optional.xml::storage-010b-make-device-short-code-optional::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
ALTER TABLE storage_device ALTER COLUMN  short_code DROP NOT NULL;

-- Set short_code to NULL for devices where code is 10 chars or less (code can be used for labels)
            -- Keep short_code for devices where code is greater than 10 chars (short_code required for labels)
            UPDATE clinlims.storage_device
            SET short_code = NULL
            WHERE LENGTH(code) <= 10;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/011-make-short-code-optional.xml::storage-010b-make-device-short-code-optional::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
