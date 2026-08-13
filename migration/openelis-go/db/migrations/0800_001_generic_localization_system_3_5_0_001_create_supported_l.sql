-- source: liquibase liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-001-create-supported-locale-seq::reagan-meant
-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE  IF NOT EXISTS clinlims.supported_locale_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS clinlims.supported_locale_seq;
-- +goose StatementEnd
