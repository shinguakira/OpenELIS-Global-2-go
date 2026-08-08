-- source: liquibase liquibase/3.4.x.x/003-create-field-tables.xml::011-003-02-create-analyzer-field-mapping-table::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Create analyzer_field_mapping table with version and denormalized analyzer_id
CREATE TABLE IF NOT EXISTS analyzer_field_mapping (id VARCHAR(36) NOT NULL, analyzer_field_id VARCHAR(36) NOT NULL, analyzer_id numeric(10, 0) NOT NULL, openelis_field_id VARCHAR(36) NOT NULL, openelis_field_type VARCHAR(20) NOT NULL, mapping_type VARCHAR(20) NOT NULL, is_required BOOLEAN DEFAULT FALSE NOT NULL, is_active BOOLEAN DEFAULT FALSE NOT NULL, specimen_type_constraint VARCHAR(255), panel_constraint VARCHAR(255), version INTEGER DEFAULT 0 NOT NULL, sys_user_id VARCHAR(36), last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT analyzer_field_mapping_pkey PRIMARY KEY (id), CONSTRAINT fk_field_mapping_analyzer_field FOREIGN KEY (analyzer_field_id) REFERENCES analyzer_field(id), CONSTRAINT fk_field_mapping_analyzer FOREIGN KEY (analyzer_id) REFERENCES analyzer(id) ON DELETE CASCADE);
ALTER TABLE clinlims.analyzer_field_mapping
            ADD CONSTRAINT chk_openelis_field_type
            CHECK (openelis_field_type IN ('TEST', 'PANEL', 'RESULT', 'ORDER', 'SAMPLE', 'QC', 'METADATA', 'UNIT'));
ALTER TABLE clinlims.analyzer_field_mapping
            ADD CONSTRAINT chk_mapping_type
            CHECK (mapping_type IN ('TEST_LEVEL', 'RESULT_LEVEL', 'METADATA'));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/003-create-field-tables.xml::011-003-02-create-analyzer-field-mapping-table::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
