-- source: liquibase liquibase/2.8.x.x/pathology.xml::2.1::csteele
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.program ADD IF NOT EXISTS manually_changed BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.program DROP COLUMN IF EXISTS manually_changed;
-- +goose StatementEnd
