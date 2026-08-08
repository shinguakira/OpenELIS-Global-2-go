-- source: liquibase liquibase/2.8.x.x/desynchronous_objects.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- create desynchronized column for provider
ALTER TABLE clinlims.provider ADD IF NOT EXISTS desynchronized BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.provider DROP COLUMN IF EXISTS desynchronized;
-- +goose StatementEnd
