-- source: liquibase liquibase/3.5.x.x/042-handling-uom-displayorder.xml::OGC-938-test-sample-handling-history-table::OGC
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.test_sample_handling_history (id VARCHAR(36) NOT NULL, test_sample_handling_id VARCHAR(36) NOT NULL, changed_by numeric(10), changed_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, change_type VARCHAR(10), previous_values JSONB, new_values JSONB, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pk_test_sample_handling_history PRIMARY KEY (id));
ALTER TABLE clinlims.test_sample_handling_history ADD CONSTRAINT fk_tsh_history_handling FOREIGN KEY (test_sample_handling_id) REFERENCES clinlims.test_sample_handling (id);
ALTER TABLE clinlims.test_sample_handling_history ADD CONSTRAINT fk_tsh_history_user FOREIGN KEY (changed_by) REFERENCES clinlims.system_user (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/042-handling-uom-displayorder.xml::OGC-938-test-sample-handling-history-table::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
