-- source: liquibase liquibase/2.3.x.x/analyzer_experiment.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE  IF NOT EXISTS clinlims.analyzer_experiment_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS clinlims.analyzer_experiment (id INTEGER NOT NULL, analyzer_id numeric(10), name VARCHAR(255) NOT NULL, file BYTEA NOT NULL, last_updated date, CONSTRAINT analyzer_experiment_pkey PRIMARY KEY (id), CONSTRAINT fk_analyzer_experiment_analyzer FOREIGN KEY (analyzer_id) REFERENCES analyzer(id));
ALTER TABLE clinlims.analyzer ADD IF NOT EXISTS has_setup_page BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/analyzer_experiment.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
