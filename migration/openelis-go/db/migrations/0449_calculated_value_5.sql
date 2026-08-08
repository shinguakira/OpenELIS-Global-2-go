-- source: liquibase liquibase/2.8.x.x/calculated_value.xml::5::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- create test_result_map table
CREATE TABLE IF NOT EXISTS test_result_map (result_calculation_id INTEGER, test_id INTEGER, result_id INTEGER);
ALTER TABLE test_result_map ADD CONSTRAINT test_result_map_result_calculation_id_fk FOREIGN KEY (result_calculation_id) REFERENCES result_calculation (id);
ALTER TABLE test_result_map ADD CONSTRAINT test_result_map_test_id_fk FOREIGN KEY (test_id) REFERENCES test (id);
ALTER TABLE test_result_map ADD CONSTRAINT test_result_map_result_id_fk FOREIGN KEY (result_id) REFERENCES result (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/calculated_value.xml::5::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
