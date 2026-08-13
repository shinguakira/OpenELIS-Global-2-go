-- source: liquibase liquibase/2.8.x.x/pathology.xml::7::csteele
-- +goose Up
-- +goose StatementBegin
-- create pathology_conclusion table
CREATE SEQUENCE  IF NOT EXISTS clinlims.pathology_conclusion_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS pathology_conclusion (id INTEGER NOT NULL, pathology_sample_id INTEGER, value VARCHAR(255), type VARCHAR(255) DEFAULT 'T', last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pathology_conclusion_pkey PRIMARY KEY (id));
ALTER TABLE pathology_conclusion ADD CONSTRAINT pathology_sample_pathology_conclusion_id_fk FOREIGN KEY (pathology_sample_id) REFERENCES pathology_sample (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/pathology.xml::7::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
