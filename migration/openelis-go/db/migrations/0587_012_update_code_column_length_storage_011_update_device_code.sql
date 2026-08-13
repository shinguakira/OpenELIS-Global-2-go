-- source: liquibase liquibase/3.3.x.x/012-update-code-column-length.xml::storage-011-update-device-code-length::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
ALTER TABLE storage_device ALTER COLUMN code TYPE VARCHAR(10) USING (code::VARCHAR(10));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/012-update-code-column-length.xml::storage-011-update-device-code-length::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
