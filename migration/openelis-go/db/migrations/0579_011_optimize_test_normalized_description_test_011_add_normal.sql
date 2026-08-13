-- source: liquibase liquibase/3.3.x.x/011-optimize-test-normalized-description.xml::test-011-add-normalized-description-column::performance-optimization
-- +goose Up
-- +goose StatementBegin
-- Add normalized_description column to improve performance of CSV test matching.
--             Normalization removes parentheses and non-alphanumeric characters for fuzzy matching.
--             Replaces inefficient in-memory normalization with database-level optimization.
ALTER TABLE test ADD IF NOT EXISTS normalized_description VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE test DROP COLUMN IF EXISTS normalized_description;
-- +goose StatementEnd
