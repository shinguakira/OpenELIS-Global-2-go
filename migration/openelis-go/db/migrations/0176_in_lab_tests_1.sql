-- source: liquibase liquibase/2.3.x.x/in_lab_tests.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.test ADD IF NOT EXISTS in_lab_only BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.test DROP COLUMN IF EXISTS in_lab_only;
-- +goose StatementEnd
