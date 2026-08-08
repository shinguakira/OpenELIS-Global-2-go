-- source: liquibase liquibase/2.1.x.x/external_connections.xml::5::csteele
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.basic_authentication_data ADD CONSTRAINT unique_basic_authentication_data_external_connection_id UNIQUE (external_connection_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.basic_authentication_data DROP COLUMN IF EXISTS CONSTRAINT;
-- +goose StatementEnd
