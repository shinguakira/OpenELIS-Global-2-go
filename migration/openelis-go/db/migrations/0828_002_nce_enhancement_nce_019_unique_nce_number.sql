-- source: liquibase liquibase/3.5.x.x/002-nce-enhancement.xml::nce-019-unique-nce-number::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add unique constraint on nce_number to prevent duplicate NCE numbers
ALTER TABLE clinlims.nc_event ADD CONSTRAINT uq_nc_event_nce_number UNIQUE (nce_number);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/002-nce-enhancement.xml::nce-019-unique-nce-number::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
