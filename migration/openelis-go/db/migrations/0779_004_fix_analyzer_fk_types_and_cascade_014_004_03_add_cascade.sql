-- source: liquibase liquibase/3.4.14.x/004-fix-analyzer-fk-types-and-cascade.xml::014-004-03-add-cascade-to-config-fks::openelis
-- +goose Up
-- +goose StatementBegin
-- Add ON DELETE CASCADE to analyzer config/setup FK tables
-- analyzer_field: config data owned by analyzer
ALTER TABLE clinlims.analyzer_field
        DROP CONSTRAINT IF EXISTS fk_analyzer_field_analyzer;

ALTER TABLE clinlims.analyzer_field
        ADD CONSTRAINT fk_analyzer_field_analyzer
        FOREIGN KEY (analyzer_id) REFERENCES clinlims.analyzer(id) ON DELETE CASCADE;

-- analyzer_experiment: experiment setup owned by analyzer
      ALTER TABLE clinlims.analyzer_experiment
        DROP CONSTRAINT IF EXISTS fk_analyzer_experiment_analyzer;

ALTER TABLE clinlims.analyzer_experiment
        ADD CONSTRAINT fk_analyzer_experiment_analyzer
        FOREIGN KEY (analyzer_id) REFERENCES clinlims.analyzer(id) ON DELETE CASCADE;

-- analyzer_pending_code: transient queue data
      ALTER TABLE clinlims.analyzer_pending_code
        DROP CONSTRAINT IF EXISTS analyzer_pending_code_analyzer_fk;

ALTER TABLE clinlims.analyzer_pending_code
        ADD CONSTRAINT analyzer_pending_code_analyzer_fk
        FOREIGN KEY (analyzer_id) REFERENCES clinlims.analyzer(id) ON DELETE CASCADE;

-- analyzer_plugin_config: config defaults/overrides
      ALTER TABLE clinlims.analyzer_plugin_config
        DROP CONSTRAINT IF EXISTS analyzer_plugin_config_analyzer_fk;

ALTER TABLE clinlims.analyzer_plugin_config
        ADD CONSTRAINT analyzer_plugin_config_analyzer_fk
        FOREIGN KEY (analyzer_id) REFERENCES clinlims.analyzer(id) ON DELETE CASCADE;

-- Verify/update tables that already have CASCADE (idempotent)
      -- analyzer_field_mapping: already CASCADE, re-assert
      ALTER TABLE clinlims.analyzer_field_mapping
        DROP CONSTRAINT IF EXISTS fk_field_mapping_analyzer;

ALTER TABLE clinlims.analyzer_field_mapping
        ADD CONSTRAINT fk_field_mapping_analyzer
        FOREIGN KEY (analyzer_id) REFERENCES clinlims.analyzer(id) ON DELETE CASCADE;

-- serial_port_configuration: already CASCADE, re-assert
      ALTER TABLE clinlims.serial_port_configuration
        DROP CONSTRAINT IF EXISTS fk_serial_port_analyzer;

ALTER TABLE clinlims.serial_port_configuration
        ADD CONSTRAINT fk_serial_port_analyzer
        FOREIGN KEY (analyzer_id) REFERENCES clinlims.analyzer(id) ON DELETE CASCADE;

-- file_import_configuration: already CASCADE, re-assert
      ALTER TABLE clinlims.file_import_configuration
        DROP CONSTRAINT IF EXISTS fk_file_import_analyzer;

ALTER TABLE clinlims.file_import_configuration
        ADD CONSTRAINT fk_file_import_analyzer
        FOREIGN KEY (analyzer_id) REFERENCES clinlims.analyzer(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/004-fix-analyzer-fk-types-and-cascade.xml::014-004-03-add-cascade-to-config-fks::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
