-- source: liquibase liquibase/3.5.x.x/shipment-002-create-shipment-table.xml::shipment-002-create-sequence::pkomena
-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE  IF NOT EXISTS shipment_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS shipment_seq;
-- +goose StatementEnd
