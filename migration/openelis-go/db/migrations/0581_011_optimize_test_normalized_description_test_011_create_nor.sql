-- source: liquibase liquibase/3.3.x.x/011-optimize-test-normalized-description.xml::test-011-create-normalized-description-index::performance-optimization
-- +goose Up
-- +goose StatementBegin
-- Create index on normalized_description for O(log n) lookup performance.
--             Replaces O(n) linear scan through all tests in memory.
CREATE INDEX IF NOT EXISTS idx_test_normalized_description ON test(normalized_description);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_test_normalized_description;
-- +goose StatementEnd
