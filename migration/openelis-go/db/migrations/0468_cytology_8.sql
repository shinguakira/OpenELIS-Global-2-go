-- source: liquibase liquibase/2.8.x.x/cytology.xml::8::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- create  cytology_diagnosis table
CREATE SEQUENCE  IF NOT EXISTS clinlims.cytology_diagnosis_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS cytology_diagnosis (id INTEGER NOT NULL, negative_diagnosis BOOLEAN, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT cytology_diagnosis_pkey PRIMARY KEY (id));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/cytology.xml::8::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
