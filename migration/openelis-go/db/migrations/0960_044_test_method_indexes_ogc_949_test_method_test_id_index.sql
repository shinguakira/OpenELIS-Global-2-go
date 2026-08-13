-- source: liquibase liquibase/3.5.x.x/044-test-method-indexes.xml::OGC-949-test-method-test-id-index::OGC
-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_test_method_test_id ON clinlims.test_method(test_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_test_method_test_id;
-- +goose StatementEnd
