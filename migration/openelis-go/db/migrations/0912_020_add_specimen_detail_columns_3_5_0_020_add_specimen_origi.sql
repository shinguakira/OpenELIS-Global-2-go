-- source: liquibase liquibase/3.5.x.x/020-add-specimen-detail-columns.xml::3.5.0-020-add-specimen-origin::claude-uat
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.sample_item ADD IF NOT EXISTS specimen_origin VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.sample_item DROP COLUMN IF EXISTS specimen_origin;
-- +goose StatementEnd
