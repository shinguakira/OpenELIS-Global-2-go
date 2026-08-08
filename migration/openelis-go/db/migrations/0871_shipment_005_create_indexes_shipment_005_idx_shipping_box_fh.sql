-- source: liquibase liquibase/3.5.x.x/shipment-005-create-indexes.xml::shipment-005-idx-shipping-box-fhir-uuid::pkomena
-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_shipping_box_fhir_uuid ON shipping_box(fhir_uuid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_shipping_box_fhir_uuid;
-- +goose StatementEnd
