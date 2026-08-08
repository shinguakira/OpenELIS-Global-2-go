-- source: liquibase liquibase/3.5.x.x/044-test-method-indexes.xml::OGC-949-test-method-unique-active::OGC
-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS uq_test_method_active ON clinlims.test_method (test_id, method_id) WHERE is_active = 'Y';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uq_test_method_active;
-- +goose StatementEnd
