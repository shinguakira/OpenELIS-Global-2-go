-- source: liquibase liquibase/2.6.x.x/provider.xml::1::ctsteele
-- +goose Up
-- +goose StatementBegin
-- add active column to provider
ALTER TABLE clinlims.provider ADD IF NOT EXISTS active BOOLEAN;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.provider DROP COLUMN IF EXISTS active;
-- +goose StatementEnd
