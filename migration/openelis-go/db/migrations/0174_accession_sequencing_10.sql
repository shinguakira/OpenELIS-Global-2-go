-- source: liquibase liquibase/2.3.x.x/accession_sequencing.xml::10::csteele
-- +goose Up
-- +goose StatementBegin
-- table for tracking accession number generation
CREATE TABLE IF NOT EXISTS clinlims.accession_number_info (prefix VARCHAR(255) NOT NULL, type VARCHAR(255) NOT NULL, cur_val BIGINT DEFAULT 1, CONSTRAINT pk_accession_number_info PRIMARY KEY (prefix, type));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clinlims.accession_number_info;
-- +goose StatementEnd
