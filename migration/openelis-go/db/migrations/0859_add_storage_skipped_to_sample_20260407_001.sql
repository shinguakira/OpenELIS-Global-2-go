-- source: liquibase liquibase/3.5.x.x/add_storage_skipped_to_sample.xml::20260407-001::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add storage_skipped field to sample table for tracking when storage assignment is intentionally skipped
ALTER TABLE clinlims.sample ADD IF NOT EXISTS storage_skipped BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.sample DROP COLUMN IF EXISTS storage_skipped;
-- +goose StatementEnd
