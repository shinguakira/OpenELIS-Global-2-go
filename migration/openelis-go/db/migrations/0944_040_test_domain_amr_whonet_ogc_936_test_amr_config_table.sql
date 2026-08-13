-- source: liquibase liquibase/3.5.x.x/040-test-domain-amr-whonet.xml::OGC-936-test-amr-config-table::OGC
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.test_amr_config (id VARCHAR(36) NOT NULL, test_id numeric(10) NOT NULL, antibiotic_code VARCHAR(20), antibiotic_class VARCHAR(255), method VARCHAR(255), breakpoint VARCHAR(255), disk_potency VARCHAR(50), lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pk_test_amr_config PRIMARY KEY (id), CONSTRAINT uq_test_amr_config_test UNIQUE (test_id));
ALTER TABLE clinlims.test_amr_config ADD CONSTRAINT fk_test_amr_config_test FOREIGN KEY (test_id) REFERENCES clinlims.test (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/040-test-domain-amr-whonet.xml::OGC-936-test-amr-config-table::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
