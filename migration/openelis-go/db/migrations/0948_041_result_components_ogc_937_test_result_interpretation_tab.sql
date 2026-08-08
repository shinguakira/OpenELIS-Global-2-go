-- source: liquibase liquibase/3.5.x.x/041-result-components.xml::OGC-937-test-result-interpretation-table::OGC
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.test_result_interpretation (id VARCHAR(36) NOT NULL, component_id VARCHAR(36) NOT NULL, value_match VARCHAR(80), interpretation_text VARCHAR(255), severity VARCHAR(20), color VARCHAR(20), display_order INTEGER DEFAULT 0 NOT NULL, is_active VARCHAR(2) DEFAULT 'Y' NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pk_test_result_interpretation PRIMARY KEY (id));
ALTER TABLE clinlims.test_result_interpretation ADD CONSTRAINT fk_test_result_interpretation_component FOREIGN KEY (component_id) REFERENCES clinlims.test_result_component (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/041-result-components.xml::OGC-937-test-result-interpretation-table::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
