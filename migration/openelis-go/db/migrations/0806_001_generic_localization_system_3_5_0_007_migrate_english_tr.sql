-- source: liquibase liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-007-migrate-english-translations::reagan-meant
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.localization_value (id, localization_id, locale, value, last_updated)
            SELECT nextval('localization_value_seq'), id, 'en', english, COALESCE(lastupdated, now())
            FROM clinlims.localization
            WHERE english IS NOT NULL AND english != '' ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-007-migrate-english-translations::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
