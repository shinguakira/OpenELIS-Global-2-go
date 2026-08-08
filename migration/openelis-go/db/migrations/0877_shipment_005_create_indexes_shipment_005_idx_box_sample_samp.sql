-- source: liquibase liquibase/3.5.x.x/shipment-005-create-indexes.xml::shipment-005-idx-box-sample-sample-id::pkomena
-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_box_sample_sample_id ON box_sample(sample_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_box_sample_sample_id;
-- +goose StatementEnd
