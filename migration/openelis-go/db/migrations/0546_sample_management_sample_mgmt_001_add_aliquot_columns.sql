-- source: liquibase liquibase/3.2.x.x/sample-management.xml::sample-mgmt-001-add-aliquot-columns::dev-team
-- +goose Up
-- +goose StatementBegin
ALTER TABLE sample_item ADD IF NOT EXISTS remaining_quantity DECIMAL(10, 3);
ALTER TABLE sample_item ADD IF NOT EXISTS parent_sample_item_id numeric(10, 0);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/sample-management.xml::sample-mgmt-001-add-aliquot-columns::dev-team
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
