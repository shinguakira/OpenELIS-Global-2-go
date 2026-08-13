-- source: liquibase liquibase/2.8.x.x/calculated_value.xml::1::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- create calculation table
CREATE SEQUENCE  IF NOT EXISTS clinlims.calculation_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS calculation (id INTEGER NOT NULL, name VARCHAR(64) NOT NULL, sample_id INTEGER NOT NULL, test_id INTEGER NOT NULL, result VARCHAR(64), toggled BOOLEAN, active BOOLEAN, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT calculation_pkey PRIMARY KEY (id));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/calculated_value.xml::1::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
