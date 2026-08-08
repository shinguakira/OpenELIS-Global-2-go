-- source: liquibase liquibase/3.3.x.x/014-rename-print-history-short-code-to-location-code.xml::storage-013-rename-print-history-short-code-to-location-code::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
ALTER TABLE storage_location_print_history RENAME COLUMN short_code TO location_code;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/014-rename-print-history-short-code-to-location-code.xml::storage-013-rename-print-history-short-code-to-location-code::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
