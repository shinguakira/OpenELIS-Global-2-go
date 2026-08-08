-- source: liquibase liquibase/3.5.x.x/004-create-site-code-sequence.xml::20260408-001::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Create sequence for auto-generating sampling site codes
CREATE SEQUENCE  IF NOT EXISTS clinlims.site_code_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS clinlims.site_code_seq;
-- +goose StatementEnd
