-- source: liquibase liquibase/3.5.x.x/shipment-012-migrate-to-sample-item.xml::3::pkomena
-- +goose Up
-- +goose StatementBegin
-- Drop old box_sample table after migration to box_sample_item
DROP TABLE IF EXISTS box_sample;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/shipment-012-migrate-to-sample-item.xml::3::pkomena
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
