-- source: liquibase liquibase/2.1.x.x/external_connections.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.external_connection (id INTEGER NOT NULL, active BOOLEAN, uri VARCHAR(255), programmed_connection VARCHAR(255), name_localization_id numeric(10), description_localization_id numeric(10), active_authentication_type VARCHAR(255), last_updated date, CONSTRAINT external_connection_pkey PRIMARY KEY (id), CONSTRAINT fk_external_connection_description_localization FOREIGN KEY (description_localization_id) REFERENCES localization(id), CONSTRAINT fk_external_connection_name_localization FOREIGN KEY (name_localization_id) REFERENCES localization(id));
CREATE SEQUENCE  IF NOT EXISTS clinlims.external_connection_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/external_connections.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
