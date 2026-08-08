-- source: liquibase liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-002-create-localization-value-seq::reagan-meant
-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE  IF NOT EXISTS clinlims.localization_value_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS clinlims.localization_value_seq;
-- +goose StatementEnd
