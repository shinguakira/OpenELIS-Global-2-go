-- source: liquibase liquibase/3.3.x.x/010-remove-odoo-site-information.xml::1::mherman22
-- +goose Up
-- +goose StatementBegin
-- Remove enableOdooConnection from site_information - moved to application.properties as org.openelisglobal.odoo.enabled
DELETE FROM clinlims.site_information WHERE name = 'enableOdooConnection';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/010-remove-odoo-site-information.xml::1::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
