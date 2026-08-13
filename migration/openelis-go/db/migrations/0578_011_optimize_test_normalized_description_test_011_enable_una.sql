-- source: liquibase liquibase/3.3.x.x/011-optimize-test-normalized-description.xml::test-011-enable-unaccent-extension::performance-optimization
-- +goose Up
-- +goose StatementBegin
-- Enable PostgreSQL unaccent extension for accent/diacritic normalization.
--             This allows UNACCENT() function to convert accented characters to their base form.
CREATE EXTENSION IF NOT EXISTS unaccent;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/011-optimize-test-normalized-description.xml::test-011-enable-unaccent-extension::performance-optimization
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
