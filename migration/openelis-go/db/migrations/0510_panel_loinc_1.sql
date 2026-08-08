-- source: liquibase liquibase/3.1.x.x/panel_loinc.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- Adds loinc columns to panel
ALTER TABLE panel ADD IF NOT EXISTS loinc VARCHAR(10);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE panel DROP COLUMN IF EXISTS loinc;
-- +goose StatementEnd
