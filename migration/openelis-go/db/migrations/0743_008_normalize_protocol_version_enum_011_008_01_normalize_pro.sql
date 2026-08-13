-- source: liquibase liquibase/3.4.x.x/008-normalize-protocol-version-enum.xml::011-008-01-normalize-protocol-version-enum::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Normalize protocol_version column to ProtocolVersion enum constant names.
--       Previously stored human-readable labels (e.g. 'ASTM LIS2-A2') or transport
--       indicators ('FILE', 'RS232'). Transport is now derived from config entities
--       (FileImportConfiguration, SerialPortConfiguration); this column only tracks
--       message format.
UPDATE analyzer SET protocol_version = 'ASTM_LIS2_A2' WHERE protocol_version IN ('ASTM LIS2-A2', 'LIS2-A2', 'ASTM')
        OR protocol_version IS NULL
        OR UPPER(protocol_version) LIKE '%FILE%'
        OR UPPER(protocol_version) LIKE '%RS232%'
        OR UPPER(protocol_version) LIKE '%RS-232%'
        OR UPPER(protocol_version) LIKE '%SERIAL%';

UPDATE analyzer SET protocol_version = 'HL7_V2_3_1' WHERE protocol_version NOT IN ('ASTM_LIS2_A2') AND UPPER(protocol_version) LIKE '%HL7%2.3%';

UPDATE analyzer SET protocol_version = 'HL7_V2_5' WHERE protocol_version NOT IN ('ASTM_LIS2_A2', 'HL7_V2_3_1') AND UPPER(protocol_version) LIKE '%HL7%2.5%';

UPDATE analyzer SET protocol_version = 'ASTM_LIS2_A2' WHERE protocol_version NOT IN ('ASTM_LIS2_A2', 'HL7_V2_3_1', 'HL7_V2_5');

ALTER TABLE analyzer ALTER COLUMN  protocol_version SET DEFAULT 'ASTM_LIS2_A2';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/008-normalize-protocol-version-enum.xml::011-008-01-normalize-protocol-version-enum::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
