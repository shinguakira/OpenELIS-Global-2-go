-- source: liquibase liquibase/3.5.x.x/043-acknowledgment-terminology.xml::OGC-939-test-terminology-mapping-table::OGC
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.test_terminology_mapping (id VARCHAR(36) NOT NULL, test_id numeric(10) NOT NULL, source VARCHAR(20) NOT NULL, code VARCHAR(80) NOT NULL, relationship VARCHAR(20), is_active VARCHAR(2) DEFAULT 'Y' NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pk_test_terminology_mapping PRIMARY KEY (id));
ALTER TABLE clinlims.test_terminology_mapping ADD CONSTRAINT uq_test_terminology_mapping UNIQUE (test_id, source, code);
ALTER TABLE clinlims.test_terminology_mapping ADD CONSTRAINT fk_test_terminology_mapping_test FOREIGN KEY (test_id) REFERENCES clinlims.test (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/043-acknowledgment-terminology.xml::OGC-939-test-terminology-mapping-table::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
