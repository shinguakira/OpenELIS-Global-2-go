-- source: liquibase liquibase/3.4.14.x/009-analyzer-discovered-source-id.xml::013-009-03::OGC-526
-- +goose Up
-- +goose StatementBegin
-- UNIQUE index on discovered_source_id for race-safe idempotency
CREATE UNIQUE INDEX IF NOT EXISTS idx_analyzer_discovered_source_id ON clinlims.analyzer(discovered_source_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_analyzer_discovered_source_id;
-- +goose StatementEnd
