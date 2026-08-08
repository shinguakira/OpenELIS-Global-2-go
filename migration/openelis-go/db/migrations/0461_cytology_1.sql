-- source: liquibase liquibase/2.8.x.x/cytology.xml::1::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- create cytology_specimen_adequacy table
CREATE SEQUENCE  IF NOT EXISTS clinlims.cytology_specimen_adequacy_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS cytology_specimen_adequacy (id INTEGER NOT NULL, result_type VARCHAR(255), satisfaction VARCHAR(255), last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT cytology_specimen_adequacy_pkey PRIMARY KEY (id));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/cytology.xml::1::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
