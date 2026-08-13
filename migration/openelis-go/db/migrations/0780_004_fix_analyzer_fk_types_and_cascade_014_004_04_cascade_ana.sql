-- source: liquibase liquibase/3.4.14.x/004-fix-analyzer-fk-types-and-cascade.xml::014-004-04-cascade-analyzer-field-children::openelis
-- +goose Up
-- +goose StatementBegin
-- Add ON DELETE CASCADE to tables referencing analyzer_field, enabling full cascade: analyzer -> analyzer_field -> children
ALTER TABLE clinlims.analyzer_field_mapping
        DROP CONSTRAINT IF EXISTS fk_field_mapping_analyzer_field;

ALTER TABLE clinlims.analyzer_field_mapping
        ADD CONSTRAINT fk_field_mapping_analyzer_field
        FOREIGN KEY (analyzer_field_id) REFERENCES clinlims.analyzer_field(id) ON DELETE CASCADE;

ALTER TABLE clinlims.qualitative_result_mapping
        DROP CONSTRAINT IF EXISTS fk_qualitative_mapping_field;

ALTER TABLE clinlims.qualitative_result_mapping
        ADD CONSTRAINT fk_qualitative_mapping_field
        FOREIGN KEY (analyzer_field_id) REFERENCES clinlims.analyzer_field(id) ON DELETE CASCADE;

ALTER TABLE clinlims.unit_mapping
        DROP CONSTRAINT IF EXISTS fk_unit_mapping_field;

ALTER TABLE clinlims.unit_mapping
        ADD CONSTRAINT fk_unit_mapping_field
        FOREIGN KEY (analyzer_field_id) REFERENCES clinlims.analyzer_field(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/004-fix-analyzer-fk-types-and-cascade.xml::014-004-04-cascade-analyzer-field-children::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
