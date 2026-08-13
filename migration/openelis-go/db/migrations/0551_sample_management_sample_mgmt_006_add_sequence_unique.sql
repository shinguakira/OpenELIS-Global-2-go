-- source: liquibase liquibase/3.2.x.x/sample-management.xml::sample-mgmt-006-add-sequence-unique::dev-team
-- +goose Up
-- +goose StatementBegin
ALTER TABLE sample_item_aliquot_relationship ADD CONSTRAINT uk_aliquot_parent_sequence UNIQUE (parent_sample_item_id, sequence_number);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/sample-management.xml::sample-mgmt-006-add-sequence-unique::dev-team
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
