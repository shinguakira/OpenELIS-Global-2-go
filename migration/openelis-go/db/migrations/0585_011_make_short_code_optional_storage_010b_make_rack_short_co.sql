-- source: liquibase liquibase/3.3.x.x/011-make-short-code-optional.xml::storage-010b-make-rack-short-code-optional::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
ALTER TABLE storage_rack ALTER COLUMN  short_code DROP NOT NULL;

-- Set short_code to NULL for racks where label is 10 chars or less (label can be used for labels)
            UPDATE clinlims.storage_rack
            SET short_code = NULL
            WHERE LENGTH(label) <= 10;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/011-make-short-code-optional.xml::storage-010b-make-rack-short-code-optional::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
