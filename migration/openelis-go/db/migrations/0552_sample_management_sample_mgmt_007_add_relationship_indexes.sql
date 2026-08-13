-- source: liquibase liquibase/3.2.x.x/sample-management.xml::sample-mgmt-007-add-relationship-indexes::dev-team
-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_aliquot_rel_parent ON sample_item_aliquot_relationship(parent_sample_item_id);
CREATE INDEX IF NOT EXISTS idx_aliquot_rel_child ON sample_item_aliquot_relationship(child_sample_item_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/sample-management.xml::sample-mgmt-007-add-relationship-indexes::dev-team
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
