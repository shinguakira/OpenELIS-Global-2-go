-- source: liquibase liquibase/3.4.x.x/009-decouple-test-mappings.xml::011-009-01-add-analyzer-type-id-column::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Add analyzer_type_id to analyzer_test_map for type-level test mappings
ALTER TABLE analyzer_test_map ADD IF NOT EXISTS analyzer_type_id numeric(10, 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE analyzer_test_map DROP COLUMN IF EXISTS analyzer_type_id;
-- +goose StatementEnd
