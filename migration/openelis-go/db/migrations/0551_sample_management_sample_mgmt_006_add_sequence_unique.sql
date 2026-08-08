-- source: liquibase liquibase/3.2.x.x/sample-management.xml::sample-mgmt-006-add-sequence-unique::dev-team
-- +goose Up
-- +goose StatementBegin
ALTER TABLE sample_item_aliquot_relationship ADD CONSTRAINT uk_aliquot_parent_sequence UNIQUE (parent_sample_item_id, sequence_number);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sample_item_aliquot_relationship DROP COLUMN IF EXISTS CONSTRAINT;
-- +goose StatementEnd
