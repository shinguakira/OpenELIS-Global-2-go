-- source: liquibase liquibase/3.5.x.x/002-nce-enhancement.xml::nce-015-add-nce-type-localization::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add name_localization_id column to nce_type table for localization support
ALTER TABLE clinlims.nce_type ADD IF NOT EXISTS name_localization_id INTEGER;
ALTER TABLE clinlims.nce_type ADD CONSTRAINT fk_nce_type_localization FOREIGN KEY (name_localization_id) REFERENCES clinlims.localization (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/002-nce-enhancement.xml::nce-015-add-nce-type-localization::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
