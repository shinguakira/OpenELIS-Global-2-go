-- source: liquibase liquibase/3.5.x.x/041-result-components.xml::OGC-937-test-result-component-table::OGC
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.test_result_component (id VARCHAR(36) NOT NULL, test_id numeric(10) NOT NULL, code VARCHAR(50) NOT NULL, label VARCHAR(255) NOT NULL, display_order INTEGER DEFAULT 0 NOT NULL, result_type VARCHAR(1), uom_id numeric(10), significant_digits numeric(10), default_result VARCHAR(80), allow_multiple_readings BOOLEAN DEFAULT FALSE NOT NULL, is_active VARCHAR(2) DEFAULT 'Y' NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pk_test_result_component PRIMARY KEY (id));
ALTER TABLE clinlims.test_result_component ADD CONSTRAINT uq_test_result_component_test_code UNIQUE (test_id, code);
ALTER TABLE clinlims.test_result_component ADD CONSTRAINT fk_test_result_component_test FOREIGN KEY (test_id) REFERENCES clinlims.test (id);
ALTER TABLE clinlims.test_result_component ADD CONSTRAINT fk_test_result_component_uom FOREIGN KEY (uom_id) REFERENCES clinlims.unit_of_measure (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/041-result-components.xml::OGC-937-test-result-component-table::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
