-- source: liquibase liquibase/2.8.x.x/reflex_rule.xml::3::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- create reflex_rule_condition table
CREATE SEQUENCE  IF NOT EXISTS clinlims.reflex_rule_condition_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS reflex_rule_condition (id INTEGER NOT NULL, sample_id VARCHAR(64) NOT NULL, test_name VARCHAR(64) NOT NULL, test_id VARCHAR(64), relation VARCHAR(64), value VARCHAR(64), value2 VARCHAR(64), reflex_rule_id INTEGER, test_analyte_id INTEGER, CONSTRAINT reflex_rule_condition_pkey PRIMARY KEY (id));
ALTER TABLE reflex_rule_condition ADD CONSTRAINT condition_reflex_rule_id_fk FOREIGN KEY (reflex_rule_id) REFERENCES reflex_rule (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/reflex_rule.xml::3::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
