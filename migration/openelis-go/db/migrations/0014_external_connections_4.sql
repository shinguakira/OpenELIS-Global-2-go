-- source: liquibase liquibase/2.1.x.x/external_connections.xml::4::csteele
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.basic_authentication_data (id INTEGER NOT NULL, external_connection_id INTEGER, username VARCHAR(255), password VARCHAR(255), last_updated date, CONSTRAINT basic_authentication_data_pkey PRIMARY KEY (id), CONSTRAINT fk_basic_authentication_data_external_connection FOREIGN KEY (external_connection_id) REFERENCES external_connection(id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clinlims.basic_authentication_data;
-- +goose StatementEnd
