-- source: liquibase liquibase/3.4.x.x/016-add-collection-conditions-column.xml::3.4.0.0-add-collection-conditions-to-sample-item::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add collection_conditions column to sample_item for storing specimen collection conditions (e.g., fasting, room temp, iced)
ALTER TABLE clinlims.sample_item ADD IF NOT EXISTS collection_conditions VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.sample_item DROP COLUMN IF EXISTS collection_conditions;
-- +goose StatementEnd
