-- source: liquibase liquibase/2.8.x.x/immunohistochemistry.xml::3::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- create immunohistochemistry_report table
CREATE SEQUENCE  IF NOT EXISTS clinlims.immunohistochemistry_report_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS immunohistochemistry_report (id INTEGER NOT NULL, immunohistochemistry_sample_id INTEGER, report_type VARCHAR(255), image BYTEA, file_type VARCHAR(255), last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT immunohistochemistry_report_pkey PRIMARY KEY (id));
ALTER TABLE immunohistochemistry_report ADD CONSTRAINT immunohistochemistry_report_immunohistochemistry_sample_id_fk FOREIGN KEY (immunohistochemistry_sample_id) REFERENCES immunohistochemistry_sample (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/immunohistochemistry.xml::3::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
