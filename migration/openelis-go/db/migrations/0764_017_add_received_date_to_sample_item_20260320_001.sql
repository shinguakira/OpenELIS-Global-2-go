-- source: liquibase liquibase/3.4.x.x/017-add-received-date-to-sample-item.xml::20260320-001::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add received_date column to sample_item table for per-sample receipt tracking
ALTER TABLE clinlims.sample_item ADD IF NOT EXISTS received_date TIMESTAMP WITH TIME ZONE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.sample_item DROP COLUMN IF EXISTS received_date;
-- +goose StatementEnd
