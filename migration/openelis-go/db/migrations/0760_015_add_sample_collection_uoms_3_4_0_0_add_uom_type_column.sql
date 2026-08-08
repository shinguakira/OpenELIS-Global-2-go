-- source: liquibase liquibase/3.4.x.x/015-add-sample-collection-uoms.xml::3.4.0.0-add-uom-type-column::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add uom_type column to unit_of_measure table
ALTER TABLE clinlims.unit_of_measure ADD IF NOT EXISTS uom_type VARCHAR(20) DEFAULT 'RESULT';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.unit_of_measure DROP COLUMN IF EXISTS uom_type;
-- +goose StatementEnd
