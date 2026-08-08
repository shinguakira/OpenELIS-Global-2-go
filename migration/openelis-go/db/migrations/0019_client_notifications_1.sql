-- source: liquibase liquibase/2.1.x.x/client_notifications.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.test ADD IF NOT EXISTS notify_results BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.test DROP COLUMN IF EXISTS notify_results;
-- +goose StatementEnd
