-- source: liquibase liquibase/3.2.x.x/sample-management.xml::sample-mgmt-002-add-parent-fk::dev-team
-- +goose Up
-- +goose StatementBegin
ALTER TABLE sample_item ADD CONSTRAINT fk_sample_item_parent FOREIGN KEY (parent_sample_item_id) REFERENCES sample_item (id) ON UPDATE CASCADE ON DELETE RESTRICT;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/sample-management.xml::sample-mgmt-002-add-parent-fk::dev-team
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
