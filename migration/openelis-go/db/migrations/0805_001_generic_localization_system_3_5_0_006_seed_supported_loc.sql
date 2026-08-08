-- source: liquibase liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-006-seed-supported-locales::reagan-meant
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.supported_locale (id, locale_code, display_name, is_active, is_fallback, sort_order, last_updated) VALUES (nextval('clinlims.supported_locale_seq'), 'en', 'English', TRUE, TRUE, 1, NOW()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.supported_locale (id, locale_code, display_name, is_active, is_fallback, sort_order, last_updated) VALUES (nextval('clinlims.supported_locale_seq'), 'fr', 'Francais', TRUE, FALSE, 2, NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-006-seed-supported-locales::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
