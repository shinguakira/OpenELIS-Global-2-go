-- source: liquibase liquibase/3.4.14.x/005-analyzer-communication-mode.xml::013-analyzer-communication-mode::pmanko
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.analyzer ADD IF NOT EXISTS communication_mode VARCHAR(25);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.analyzer DROP COLUMN IF EXISTS communication_mode;
-- +goose StatementEnd
