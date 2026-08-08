-- source: liquibase liquibase/2.6.x.x/add_order_priority.xml::2::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- add Priority column to electronic_order table
ALTER TABLE clinlims.electronic_order ADD IF NOT EXISTS order_priority VARCHAR(255) DEFAULT 'ROUTINE';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.electronic_order DROP COLUMN IF EXISTS order_priority;
-- +goose StatementEnd
