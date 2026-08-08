-- source: liquibase liquibase/3.3.x.x/029-label-preset-tables.xml::label-preset-003-test-label-config::ogc-285
-- +goose Up
-- +goose StatementBegin
-- Create test_label_config table: 1:1 master label toggle per test (FRS §7.2)
CREATE SEQUENCE  IF NOT EXISTS clinlims.test_label_config_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS clinlims.test_label_config (id INTEGER DEFAULT nextval('test_label_config_seq') NOT NULL, test_id numeric(10, 0) NOT NULL, allow_order_entry_override BOOLEAN DEFAULT TRUE NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT test_label_config_pkey PRIMARY KEY (id), CONSTRAINT fk_test_label_config_test FOREIGN KEY (test_id) REFERENCES test(id) ON DELETE CASCADE, UNIQUE (test_id));
CREATE INDEX IF NOT EXISTS idx_test_label_config_test_id ON clinlims.test_label_config(test_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/029-label-preset-tables.xml::label-preset-003-test-label-config::ogc-285
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
