-- source: liquibase liquibase/3.5.x.x/039-test-method-links.xml::OGC-746-test-method-seq::OGC
-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE  IF NOT EXISTS clinlims.test_method_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS clinlims.test_method_seq;
-- +goose StatementEnd
