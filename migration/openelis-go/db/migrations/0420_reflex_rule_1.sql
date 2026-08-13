-- source: liquibase liquibase/2.8.x.x/reflex_rule.xml::1::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- create reflex_rule table
CREATE SEQUENCE  IF NOT EXISTS clinlims.reflex_rule_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS reflex_rule (id INTEGER NOT NULL, rule_name VARCHAR(64) NOT NULL, overall VARCHAR(64) NOT NULL, toggled BOOLEAN, active BOOLEAN, last_updated TIMESTAMP WITHOUT TIME ZONE, analyte_id INTEGER, CONSTRAINT reflex_rule_pkey PRIMARY KEY (id));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/reflex_rule.xml::1::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
