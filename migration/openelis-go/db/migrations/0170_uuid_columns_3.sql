-- source: liquibase liquibase/2.3.x.x/uuid_columns.xml::3::csteele
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.electronic_order ADD IF NOT EXISTS type VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.electronic_order DROP COLUMN IF EXISTS type;
-- +goose StatementEnd
