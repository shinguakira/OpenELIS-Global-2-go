-- source: liquibase liquibase/3.3.x.x/008-add-short-code-columns.xml::storage-008-add-short-code-to-device::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
ALTER TABLE storage_device ADD IF NOT EXISTS short_code VARCHAR(10);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE storage_device DROP COLUMN IF EXISTS short_code;
-- +goose StatementEnd
