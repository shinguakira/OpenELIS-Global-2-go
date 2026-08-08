-- source: liquibase liquibase/2.6.x.x/add_order_priority.xml::1::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- add Priority column to sample table
ALTER TABLE clinlims.sample ADD IF NOT EXISTS order_priority VARCHAR(255) DEFAULT 'ROUTINE';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.sample DROP COLUMN IF EXISTS order_priority;
-- +goose StatementEnd
