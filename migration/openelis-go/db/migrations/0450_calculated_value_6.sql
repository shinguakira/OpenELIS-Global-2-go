-- source: liquibase liquibase/2.8.x.x/calculated_value.xml::6::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- add result_calculated  Column to analysis table
ALTER TABLE clinlims.analysis ADD IF NOT EXISTS result_calculated BOOLEAN;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.analysis DROP COLUMN IF EXISTS result_calculated;
-- +goose StatementEnd
