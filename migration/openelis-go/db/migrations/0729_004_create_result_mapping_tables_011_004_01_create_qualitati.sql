-- source: liquibase liquibase/3.4.x.x/004-create-result-mapping-tables.xml::011-004-01-create-qualitative-result-mapping-table::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Create qualitative_result_mapping table for analyzer-to-OpenELIS value mappings
CREATE TABLE IF NOT EXISTS qualitative_result_mapping (id VARCHAR(36) NOT NULL, analyzer_field_id VARCHAR(36) NOT NULL, analyzer_value VARCHAR(100) NOT NULL, openelis_code VARCHAR(100) NOT NULL, is_default BOOLEAN DEFAULT FALSE NOT NULL, sys_user_id VARCHAR(36), last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT qualitative_result_mapping_pkey PRIMARY KEY (id), CONSTRAINT fk_qualitative_mapping_field FOREIGN KEY (analyzer_field_id) REFERENCES analyzer_field(id));
ALTER TABLE qualitative_result_mapping ADD CONSTRAINT uk_qualitative_mapping_field_value UNIQUE (analyzer_field_id, analyzer_value);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/004-create-result-mapping-tables.xml::011-004-01-create-qualitative-result-mapping-table::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
