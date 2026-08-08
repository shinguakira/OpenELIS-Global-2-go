-- source: liquibase liquibase/2.0.x.x/validation_site_information.xml::1::caleb
-- +goose Up
-- +goose StatementBegin
-- add lastupdated to site informations without to stop a bug
UPDATE clinlims.site_information SET lastupdated = now() WHERE lastupdated IS NULL;

ALTER TABLE clinlims.site_information ALTER COLUMN  lastupdated SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.0.x.x/validation_site_information.xml::1::caleb
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
