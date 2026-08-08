-- source: liquibase liquibase/3.2.x.x/notebook.xml::create-notebook-file-table::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Creating NOTEBOOK_FILE table to store binary file attachments related to a notebook.
CREATE TABLE IF NOT EXISTS notebook_file (id INTEGER NOT NULL, file_data BYTEA, file_type VARCHAR(255), file_name VARCHAR(255), last_updated TIMESTAMP WITHOUT TIME ZONE, notebook_id INTEGER NOT NULL, CONSTRAINT pk_notebook_file PRIMARY KEY (id), CONSTRAINT fk_file_notebook FOREIGN KEY (notebook_id) REFERENCES notebook(id));
CREATE SEQUENCE  IF NOT EXISTS notebook_file_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/notebook.xml::create-notebook-file-table::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
