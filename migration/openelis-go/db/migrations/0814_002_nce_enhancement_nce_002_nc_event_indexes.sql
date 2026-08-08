-- source: liquibase liquibase/3.5.x.x/002-nce-enhancement.xml::nce-002-nc-event-indexes::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add indexes for nc_event columns
CREATE INDEX IF NOT EXISTS idx_nc_event_status ON clinlims.nc_event(status);
CREATE INDEX IF NOT EXISTS idx_nc_event_severity ON clinlims.nc_event(severity);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/002-nce-enhancement.xml::nce-002-nc-event-indexes::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
