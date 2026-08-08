-- source: liquibase liquibase/3.5.x.x/024-migrate-madagascar-address-fields-to-identities.xml::3.5.0-024-address-hierarchy-identity-types::openelisglobal
-- +goose Up
-- +goose StatementBegin
-- Create generic identity types for address hierarchy values before migrating off Madagascar-specific person columns.
INSERT INTO clinlims.patient_identity_type (id, identity_type, description, lastupdated)
            SELECT nextval('clinlims.patient_identity_type_seq'), 'ADDRESS_HIERARCHY_0', 'Address hierarchy level 1', now()
            WHERE NOT EXISTS (
                SELECT 1 FROM clinlims.patient_identity_type WHERE identity_type = 'ADDRESS_HIERARCHY_0'
            ) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.patient_identity_type (id, identity_type, description, lastupdated)
            SELECT nextval('clinlims.patient_identity_type_seq'), 'ADDRESS_HIERARCHY_3', 'Address hierarchy level 4', now()
            WHERE NOT EXISTS (
                SELECT 1 FROM clinlims.patient_identity_type WHERE identity_type = 'ADDRESS_HIERARCHY_3'
            ) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.patient_identity_type (id, identity_type, description, lastupdated)
            SELECT nextval('clinlims.patient_identity_type_seq'), 'ADDRESS_HIERARCHY_4', 'Address hierarchy level 5', now()
            WHERE NOT EXISTS (
                SELECT 1 FROM clinlims.patient_identity_type WHERE identity_type = 'ADDRESS_HIERARCHY_4'
            ) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/024-migrate-madagascar-address-fields-to-identities.xml::3.5.0-024-address-hierarchy-identity-types::openelisglobal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
