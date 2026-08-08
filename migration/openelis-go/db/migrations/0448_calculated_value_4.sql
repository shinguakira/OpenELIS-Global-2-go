-- source: liquibase liquibase/2.8.x.x/calculated_value.xml::4::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- create test_operations table
CREATE TABLE IF NOT EXISTS test_operations (result_calculation_id INTEGER, test_id INTEGER);
ALTER TABLE test_operations ADD CONSTRAINT test_opeartions_result_calculation_id_fk FOREIGN KEY (result_calculation_id) REFERENCES result_calculation (id);
ALTER TABLE test_operations ADD CONSTRAINT test_opeartions_test_id_fk FOREIGN KEY (test_id) REFERENCES test (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/calculated_value.xml::4::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
