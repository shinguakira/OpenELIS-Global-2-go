-- source: liquibase liquibase/3.1.x.x/dictionary-loinc.xml::add-dictionary-loinc-code::agaba-derrick
-- +goose Up
-- +goose StatementBegin
-- Adds LOINC code column to dictionary for result mappings
ALTER TABLE dictionary ADD IF NOT EXISTS loinc_code VARCHAR(10);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE dictionary DROP COLUMN IF EXISTS loinc_code;
-- +goose StatementEnd
