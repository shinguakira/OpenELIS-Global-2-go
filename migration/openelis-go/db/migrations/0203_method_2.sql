-- source: liquibase liquibase/2.6.x.x/method.xml::2::cliff
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.method ADD IF NOT EXISTS name_localization_id numeric(10);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.method DROP COLUMN IF EXISTS name_localization_id;
-- +goose StatementEnd
