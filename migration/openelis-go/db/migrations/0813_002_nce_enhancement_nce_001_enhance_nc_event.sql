-- source: liquibase liquibase/3.5.x.x/002-nce-enhancement.xml::nce-001-enhance-nc-event::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add title, immediate_action, and severity columns to nc_event
ALTER TABLE clinlims.nc_event ADD IF NOT EXISTS title VARCHAR(200);
ALTER TABLE clinlims.nc_event ADD IF NOT EXISTS immediate_action TEXT;
ALTER TABLE clinlims.nc_event ADD IF NOT EXISTS severity VARCHAR(20);
ALTER TABLE clinlims.nc_event ADD IF NOT EXISTS last_updated TIMESTAMP WITHOUT TIME ZONE;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/002-nce-enhancement.xml::nce-001-enhance-nc-event::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
