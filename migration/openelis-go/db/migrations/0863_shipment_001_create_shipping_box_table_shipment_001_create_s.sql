-- source: liquibase liquibase/3.5.x.x/shipment-001-create-shipping-box-table.xml::shipment-001-create-shipping-box-table::pkomena
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS shipping_box (id INTEGER NOT NULL, box_id VARCHAR(50) NOT NULL, fhir_uuid UUID NOT NULL, destination_facility_id INTEGER NOT NULL, state VARCHAR(50) NOT NULL, temperature_requirement VARCHAR(50), notes TEXT, created_date TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, created_by INTEGER, sent_date TIMESTAMP WITHOUT TIME ZONE, received_date TIMESTAMP WITHOUT TIME ZONE, reconciled_date TIMESTAMP WITHOUT TIME ZONE, archived BOOLEAN DEFAULT FALSE NOT NULL, archived_date TIMESTAMP WITHOUT TIME ZONE, sys_user_id INTEGER NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT shipping_box_pkey PRIMARY KEY (id), UNIQUE (box_id), UNIQUE (fhir_uuid));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS shipping_box;
-- +goose StatementEnd
