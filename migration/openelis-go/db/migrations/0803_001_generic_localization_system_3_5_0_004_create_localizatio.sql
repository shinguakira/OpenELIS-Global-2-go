-- source: liquibase liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-004-create-localization-value-table::reagan-meant
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.localization_value (id numeric NOT NULL, localization_id numeric NOT NULL, locale VARCHAR(10) NOT NULL, value TEXT NOT NULL, last_updated TIMESTAMP with time zone, CONSTRAINT localization_value_pkey PRIMARY KEY (id), CONSTRAINT fk_localization_value_localization FOREIGN KEY (localization_id) REFERENCES clinlims.localization(id) ON DELETE CASCADE);
ALTER TABLE clinlims.localization_value ADD CONSTRAINT uq_localization_value_locale UNIQUE (localization_id, locale);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-004-create-localization-value-table::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
