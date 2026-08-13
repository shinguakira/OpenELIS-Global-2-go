-- source: liquibase liquibase/3.5.x.x/020-add-specimen-detail-columns.xml::3.5.0-020-add-sample-temperature::claude-uat
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.sample_item ADD IF NOT EXISTS sample_temperature VARCHAR(50);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.sample_item DROP COLUMN IF EXISTS sample_temperature;
-- +goose StatementEnd
