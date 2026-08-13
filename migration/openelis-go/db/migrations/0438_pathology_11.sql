-- source: liquibase liquibase/2.8.x.x/pathology.xml::11::csteele
-- +goose Up
-- +goose StatementBegin
ALTER TABLE pathology_slide ADD IF NOT EXISTS location VARCHAR(255);
ALTER TABLE pathology_block ADD IF NOT EXISTS location VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/pathology.xml::11::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
