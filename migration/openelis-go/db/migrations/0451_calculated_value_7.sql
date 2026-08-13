-- source: liquibase liquibase/2.8.x.x/calculated_value.xml::7::mozzymutesa
-- +goose Up
-- +goose StatementBegin
ALTER TABLE calculation ADD CONSTRAINT calculation_name_unique_constraint UNIQUE (name);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/calculated_value.xml::7::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
