-- source: liquibase liquibase/3.5.x.x/shipment-005-create-indexes.xml::shipment-005-idx-shipping-box-state::pkomena
-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_shipping_box_state ON shipping_box(state);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_shipping_box_state;
-- +goose StatementEnd
