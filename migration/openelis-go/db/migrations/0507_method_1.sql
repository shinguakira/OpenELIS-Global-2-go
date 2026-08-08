-- source: liquibase liquibase/2.8.x.x/method.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- Add Method to analysis
ALTER TABLE clinlims.analysis ADD IF NOT EXISTS method_id numeric(10);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.analysis DROP COLUMN IF EXISTS method_id;
-- +goose StatementEnd
