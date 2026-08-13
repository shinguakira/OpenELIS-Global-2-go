-- source: liquibase liquibase/3.4.x.x/001-create-standalone-tables.xml::011-001-02-create-analyzer-type-sequence::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Create sequence for analyzer_type primary key generation
CREATE SEQUENCE  IF NOT EXISTS analyzer_type_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS analyzer_type_seq;
-- +goose StatementEnd
