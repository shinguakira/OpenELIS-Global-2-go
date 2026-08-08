-- source: liquibase liquibase/3.3.x.x/eqa-011-add-enrollment-program-name.xml::eqa-011-05-add-provider-to-eqa-program::mozzy11
-- +goose Up
-- +goose StatementBegin
-- Add free-text provider column to eqa_program table
ALTER TABLE clinlims.eqa_program ADD IF NOT EXISTS provider VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.eqa_program DROP COLUMN IF EXISTS provider;
-- +goose StatementEnd
