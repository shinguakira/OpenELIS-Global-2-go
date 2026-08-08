-- source: liquibase liquibase/3.5.x.x/042-handling-uom-displayorder.xml::OGC-938-test-sample-handling-table::OGC
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.test_sample_handling (id VARCHAR(36) NOT NULL, test_id numeric(10) NOT NULL, storage_condition VARCHAR(50), storage_condition_custom VARCHAR(200), storage_duration INTEGER, storage_duration_unit VARCHAR(20), stability_notes TEXT, protect_from_light BOOLEAN DEFAULT FALSE NOT NULL, do_not_freeze BOOLEAN DEFAULT FALSE NOT NULL, do_not_refrigerate BOOLEAN DEFAULT FALSE NOT NULL, disposal_method VARCHAR(100), disposal_timeframe INTEGER, disposal_unit VARCHAR(20), special_instructions TEXT, override_restricted BOOLEAN DEFAULT FALSE NOT NULL, version INTEGER DEFAULT 0 NOT NULL, is_active VARCHAR(2) DEFAULT 'Y' NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pk_test_sample_handling PRIMARY KEY (id), CONSTRAINT uq_test_sample_handling_test UNIQUE (test_id));
ALTER TABLE clinlims.test_sample_handling ADD CONSTRAINT fk_test_sample_handling_test FOREIGN KEY (test_id) REFERENCES clinlims.test (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/042-handling-uom-displayorder.xml::OGC-938-test-sample-handling-table::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
