-- source: liquibase liquibase/3.2.x.x/sample-management.xml::sample-mgmt-008-migrate-legacy-quantity::dev-team
-- +goose Up
-- +goose StatementBegin
UPDATE sample_item
            SET remaining_quantity = quantity
            WHERE quantity IS NOT NULL
              AND remaining_quantity IS NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/sample-management.xml::sample-mgmt-008-migrate-legacy-quantity::dev-team
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
