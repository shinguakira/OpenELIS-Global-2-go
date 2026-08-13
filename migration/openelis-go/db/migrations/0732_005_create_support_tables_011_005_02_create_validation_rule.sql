-- source: liquibase liquibase/3.4.x.x/005-create-support-tables.xml::011-005-02-create-validation-rule-configuration-table::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Create validation_rule_configuration table for custom field type validation
CREATE TABLE IF NOT EXISTS validation_rule_configuration (id VARCHAR(36) NOT NULL, custom_field_type_id VARCHAR(36) NOT NULL, rule_name VARCHAR(100) NOT NULL, rule_type VARCHAR(20) NOT NULL, rule_expression TEXT, error_message VARCHAR(500), is_active BOOLEAN DEFAULT TRUE NOT NULL, sys_user_id VARCHAR(36), last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT validation_rule_configuration_pkey PRIMARY KEY (id), CONSTRAINT fk_validation_rule_custom_type FOREIGN KEY (custom_field_type_id) REFERENCES custom_field_type(id) ON DELETE CASCADE);
ALTER TABLE clinlims.validation_rule_configuration
            ADD CONSTRAINT chk_rule_type
            CHECK (rule_type IN ('REGEX', 'RANGE', 'ENUM', 'LENGTH'));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/005-create-support-tables.xml::011-005-02-create-validation-rule-configuration-table::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
