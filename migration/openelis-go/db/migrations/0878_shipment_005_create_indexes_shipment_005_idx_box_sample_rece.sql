-- source: liquibase liquibase/3.5.x.x/shipment-005-create-indexes.xml::shipment-005-idx-box-sample-reception-status::pkomena
-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_box_sample_reception_status ON box_sample(reception_status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_box_sample_reception_status;
-- +goose StatementEnd
