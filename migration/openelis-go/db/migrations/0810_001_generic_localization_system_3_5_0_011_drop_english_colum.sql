-- source: liquibase liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-011-drop-english-column::reagan-meant
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.localization DROP COLUMN IF EXISTS english;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-011-drop-english-column::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
