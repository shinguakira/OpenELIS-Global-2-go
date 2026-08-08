-- source: liquibase liquibase/3.4.14.x/009-analyzer-discovered-source-id.xml::013-009-01::OGC-526
-- +goose Up
-- +goose StatementBegin
-- Add discovered_source_id column for bridge-discovered analyzer stubs
ALTER TABLE clinlims.analyzer ADD IF NOT EXISTS discovered_source_id VARCHAR(500);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.analyzer DROP COLUMN IF EXISTS discovered_source_id;
-- +goose StatementEnd
