-- source: liquibase liquibase/2.8.x.x/reflex_rule.xml::2::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- create reflex_rule_action table
CREATE SEQUENCE  IF NOT EXISTS clinlims.reflex_rule_action_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS reflex_rule_action (id INTEGER NOT NULL, reflex_test_name VARCHAR(64) NOT NULL, reflex_test_id VARCHAR(64) NOT NULL, sample_id VARCHAR(64) NOT NULL, internal_note VARCHAR(64), external_note VARCHAR(64), add_notification VARCHAR(2), reflex_rule_id INTEGER, test_reflex_id INTEGER, CONSTRAINT reflex_rule_action_pkey PRIMARY KEY (id));
ALTER TABLE reflex_rule_action ADD CONSTRAINT action_reflex_rule_id_fk FOREIGN KEY (reflex_rule_id) REFERENCES reflex_rule (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/reflex_rule.xml::2::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
