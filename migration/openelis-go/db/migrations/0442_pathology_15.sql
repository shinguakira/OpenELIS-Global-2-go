-- source: liquibase liquibase/2.8.x.x/pathology.xml::15::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- create pathology_report table
CREATE SEQUENCE  IF NOT EXISTS clinlims.pathology_report_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS pathology_report (id INTEGER NOT NULL, pathology_sample_id INTEGER, report_type VARCHAR(255), image BYTEA, file_type VARCHAR(255), last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pathology_report_pkey PRIMARY KEY (id));
ALTER TABLE pathology_report ADD CONSTRAINT pathology_report_pathology_sample_id_fk FOREIGN KEY (pathology_sample_id) REFERENCES pathology_sample (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/pathology.xml::15::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
