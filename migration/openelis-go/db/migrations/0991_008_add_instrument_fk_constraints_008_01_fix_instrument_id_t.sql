-- source: liquibase liquibase/qc/008-add-instrument-fk-constraints.xml::008-01-fix-instrument-id-types::code-review
-- +goose Up
-- +goose StatementBegin
-- Align instrument_id columns from INT to NUMERIC(10,0) to match analyzer.id type
ALTER TABLE qc_control_lot ALTER COLUMN instrument_id TYPE numeric(10, 0) USING (instrument_id::numeric(10, 0));

ALTER TABLE qc_result ALTER COLUMN instrument_id TYPE numeric(10, 0) USING (instrument_id::numeric(10, 0));

ALTER TABLE westgard_rule_config ALTER COLUMN instrument_id TYPE numeric(10, 0) USING (instrument_id::numeric(10, 0));

ALTER TABLE qc_rule_violation ALTER COLUMN instrument_id TYPE numeric(10, 0) USING (instrument_id::numeric(10, 0));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/qc/008-add-instrument-fk-constraints.xml::008-01-fix-instrument-id-types::code-review
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
