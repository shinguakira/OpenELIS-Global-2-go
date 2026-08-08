-- source: liquibase liquibase/3.5.x.x/shipment-002-create-shipment-table.xml::shipment-002-create-shipment-table::pkomena
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS shipment (id INTEGER NOT NULL, shipping_box_id INTEGER NOT NULL, courier VARCHAR(100), tracking_number VARCHAR(100), shipped_date TIMESTAMP WITHOUT TIME ZONE, estimated_delivery_date TIMESTAMP WITHOUT TIME ZONE, actual_delivery_date TIMESTAMP WITHOUT TIME ZONE, sender_notes TEXT, receiver_notes TEXT, status VARCHAR(50) NOT NULL, sys_user_id INTEGER NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT shipment_pkey PRIMARY KEY (id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS shipment;
-- +goose StatementEnd
