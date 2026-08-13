-- source: liquibase liquibase/qc/008-add-instrument-fk-constraints.xml::008-02-add-instrument-fk-constraints::code-review
-- +goose Up
-- +goose StatementBegin
-- Add missing foreign key constraints for instrument_id columns referencing analyzer table
ALTER TABLE qc_control_lot ADD CONSTRAINT fk_qc_control_lot_analyzer FOREIGN KEY (instrument_id) REFERENCES analyzer (id);

ALTER TABLE qc_result ADD CONSTRAINT fk_qc_result_analyzer FOREIGN KEY (instrument_id) REFERENCES analyzer (id);

ALTER TABLE westgard_rule_config ADD CONSTRAINT fk_westgard_rule_config_analyzer FOREIGN KEY (instrument_id) REFERENCES analyzer (id);

ALTER TABLE qc_rule_violation ADD CONSTRAINT fk_qc_rule_violation_analyzer FOREIGN KEY (instrument_id) REFERENCES analyzer (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/qc/008-add-instrument-fk-constraints.xml::008-02-add-instrument-fk-constraints::code-review
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
