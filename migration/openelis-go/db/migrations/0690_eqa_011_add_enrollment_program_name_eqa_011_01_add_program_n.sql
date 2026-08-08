-- source: liquibase liquibase/3.3.x.x/eqa-011-add-enrollment-program-name.xml::eqa-011-01-add-program-name-column::mozzy11
-- +goose Up
-- +goose StatementBegin
-- Add program_name column to self-enrollment table for free-text program names
ALTER TABLE clinlims.eqa_lab_program_enrollment ADD IF NOT EXISTS program_name VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.eqa_lab_program_enrollment DROP COLUMN IF EXISTS program_name;
-- +goose StatementEnd
