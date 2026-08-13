-- source: liquibase liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-003-create-supported-locale-table::reagan-meant
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.supported_locale (id numeric NOT NULL, locale_code VARCHAR(10) NOT NULL, display_name TEXT NOT NULL, is_active BOOLEAN DEFAULT TRUE NOT NULL, is_fallback BOOLEAN DEFAULT FALSE NOT NULL, sort_order INTEGER DEFAULT 0 NOT NULL, last_updated TIMESTAMP with time zone, CONSTRAINT supported_locale_pkey PRIMARY KEY (id), UNIQUE (locale_code));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clinlims.supported_locale;
-- +goose StatementEnd
