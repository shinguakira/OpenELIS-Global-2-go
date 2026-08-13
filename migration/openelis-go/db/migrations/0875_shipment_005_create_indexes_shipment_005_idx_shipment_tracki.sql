-- source: liquibase liquibase/3.5.x.x/shipment-005-create-indexes.xml::shipment-005-idx-shipment-tracking::pkomena
-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_shipment_tracking ON shipment(tracking_number);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_shipment_tracking;
-- +goose StatementEnd
