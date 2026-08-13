-- source: liquibase liquibase/3.4.14.x/001-add-file-format-to-file-import-config.xml::014-001-add-file-format-to-file-import-config::openelis
-- +goose Up
-- +goose StatementBegin
ALTER TABLE file_import_configuration ADD IF NOT EXISTS file_format VARCHAR(20) DEFAULT 'CSV' NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE file_import_configuration DROP COLUMN IF EXISTS file_format;
-- +goose StatementEnd
