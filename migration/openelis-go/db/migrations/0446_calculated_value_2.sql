-- source: liquibase liquibase/2.8.x.x/calculated_value.xml::2::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- create calculation_operation table
CREATE SEQUENCE  IF NOT EXISTS clinlims.calculation_operation_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS calculation_operation (id INTEGER NOT NULL, type VARCHAR(64) NOT NULL, sample_id INTEGER, operation_order INTEGER, value VARCHAR(64) NOT NULL, calculation_id INTEGER, CONSTRAINT calculation_operation_pkey PRIMARY KEY (id));
ALTER TABLE calculation_operation ADD CONSTRAINT calculation_operation_calculation_id_fk FOREIGN KEY (calculation_id) REFERENCES calculation (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/calculated_value.xml::2::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
