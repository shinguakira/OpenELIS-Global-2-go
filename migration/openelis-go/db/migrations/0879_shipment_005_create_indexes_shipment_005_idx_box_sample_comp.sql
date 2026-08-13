-- source: liquibase liquibase/3.5.x.x/shipment-005-create-indexes.xml::shipment-005-idx-box-sample-composite::pkomena
-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS idx_box_sample_composite ON box_sample(shipping_box_id, sample_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_box_sample_composite;
-- +goose StatementEnd
