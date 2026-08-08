-- source: liquibase liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-005-create-localization-value-indexes::reagan-meant
-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_localization_value_locale ON clinlims.localization_value(locale);
CREATE INDEX IF NOT EXISTS idx_localization_value_localization_id ON clinlims.localization_value(localization_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-005-create-localization-value-indexes::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
