-- source: liquibase liquibase/2.1.x.x/external_connections.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.external_connection_contact (id INTEGER NOT NULL, external_connection_id INTEGER, person_id numeric(10), last_updated date, CONSTRAINT external_connection_contact_pkey PRIMARY KEY (id), CONSTRAINT fk_external_connection_contact_external_connection FOREIGN KEY (external_connection_id) REFERENCES external_connection(id), CONSTRAINT fk_external_connection_contact_person FOREIGN KEY (person_id) REFERENCES person(id));
CREATE SEQUENCE  IF NOT EXISTS clinlims.external_connection_contact_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/external_connections.xml::2::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
