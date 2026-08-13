-- source: liquibase liquibase/3.5.x.x/004-analysis-analyzer-id.xml::3.5.0-003-add-analysis-analyzer-id-column::pmanko
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.analysis ADD IF NOT EXISTS analyzer_id numeric(10);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.analysis DROP COLUMN IF EXISTS analyzer_id;
-- +goose StatementEnd
