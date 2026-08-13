-- source: liquibase liquibase/2.8.x.x/add_antimicrobial_resistance_test_column.xml::1::rossumg
-- +goose Up
-- +goose StatementBegin
-- Add Antimicrobial Resistance column to test table
ALTER TABLE test ADD IF NOT EXISTS antimicrobial_resistance BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE test DROP COLUMN IF EXISTS antimicrobial_resistance;
-- +goose StatementEnd
