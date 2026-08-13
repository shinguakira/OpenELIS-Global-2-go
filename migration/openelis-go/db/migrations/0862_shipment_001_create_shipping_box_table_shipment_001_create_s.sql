-- source: liquibase liquibase/3.5.x.x/shipment-001-create-shipping-box-table.xml::shipment-001-create-sequence::pkomena
-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE  IF NOT EXISTS shipping_box_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS shipping_box_seq;
-- +goose StatementEnd
