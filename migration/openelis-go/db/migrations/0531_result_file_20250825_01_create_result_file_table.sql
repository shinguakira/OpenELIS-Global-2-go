-- source: liquibase liquibase/3.2.x.x/result_file.xml::20250825-01-create-result-file-table::elia
-- +goose Up
-- +goose StatementBegin
-- Creating the result_file table to store result files
CREATE SEQUENCE  IF NOT EXISTS result_file_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS result_file (id INTEGER DEFAULT nextval('result_file_seq') NOT NULL, file_name VARCHAR(255) NOT NULL, file_type VARCHAR(100) NOT NULL, file_content BYTEA NOT NULL, uploaded_at TIMESTAMP WITHOUT TIME ZONE NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE NOT NULL, CONSTRAINT pk_result_file_id PRIMARY KEY (id), UNIQUE (id));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/result_file.xml::20250825-01-create-result-file-table::elia
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
