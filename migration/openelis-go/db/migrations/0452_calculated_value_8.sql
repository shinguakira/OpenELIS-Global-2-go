-- source: liquibase liquibase/2.8.x.x/calculated_value.xml::8::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- add note column to calculation table
ALTER TABLE clinlims.calculation ADD IF NOT EXISTS note VARCHAR(64);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.calculation DROP COLUMN IF EXISTS note;
-- +goose StatementEnd
