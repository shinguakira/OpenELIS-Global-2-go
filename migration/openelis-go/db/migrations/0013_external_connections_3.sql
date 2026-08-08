-- source: liquibase liquibase/2.1.x.x/external_connections.xml::3::csteele
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.certificate_authentication_data (id INTEGER NOT NULL, external_connection_id INTEGER, last_updated date, CONSTRAINT certificate_authentication_data_pkey PRIMARY KEY (id), CONSTRAINT fk_certificate_authentication_data_external_connection FOREIGN KEY (external_connection_id) REFERENCES external_connection(id));
CREATE SEQUENCE  IF NOT EXISTS clinlims.external_connection_authentication_data_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/external_connections.xml::3::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
