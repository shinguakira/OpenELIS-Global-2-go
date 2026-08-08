-- source: liquibase liquibase/3.4.x.x/004-create-result-mapping-tables.xml::011-004-02-create-unit-mapping-table::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Create unit_mapping table for analyzer-to-OpenELIS unit conversion
CREATE TABLE IF NOT EXISTS unit_mapping (id VARCHAR(36) NOT NULL, analyzer_field_id VARCHAR(36) NOT NULL, analyzer_unit VARCHAR(50) NOT NULL, openelis_unit VARCHAR(50) NOT NULL, conversion_factor DECIMAL(10, 6), reject_if_mismatch BOOLEAN DEFAULT FALSE NOT NULL, sys_user_id VARCHAR(36), last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT unit_mapping_pkey PRIMARY KEY (id), CONSTRAINT fk_unit_mapping_field FOREIGN KEY (analyzer_field_id) REFERENCES analyzer_field(id));
ALTER TABLE unit_mapping ADD CONSTRAINT uk_unit_mapping_field_unit UNIQUE (analyzer_field_id, analyzer_unit);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/004-create-result-mapping-tables.xml::011-004-02-create-unit-mapping-table::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
