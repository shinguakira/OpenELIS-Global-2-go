-- source: liquibase liquibase/3.5.x.x/shipment-003-create-box-sample-table.xml::shipment-003-create-sequence::pkomena
-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE  IF NOT EXISTS box_sample_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS box_sample_seq;
-- +goose StatementEnd
