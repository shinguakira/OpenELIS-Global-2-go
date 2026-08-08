-- source: liquibase liquibase/2.0.x.x/default_test.xml::1::caleb
-- +goose Up
-- +goose StatementBegin
-- add lastupdated to site informations without to stop a bug
ALTER TABLE clinlims.test ADD IF NOT EXISTS default_test_result_id numeric(10);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.test DROP COLUMN IF EXISTS default_test_result_id;
-- +goose StatementEnd
