-- source: liquibase liquibase/3.5.x.x/shipment-003-create-box-sample-table.xml::shipment-003-create-box-sample-table::pkomena
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS box_sample (id INTEGER NOT NULL, shipping_box_id INTEGER NOT NULL, sample_id INTEGER NOT NULL, added_date TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, position_in_box INTEGER, reception_status VARCHAR(50), reception_notes TEXT, sys_user_id INTEGER NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT box_sample_pkey PRIMARY KEY (id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS box_sample;
-- +goose StatementEnd
