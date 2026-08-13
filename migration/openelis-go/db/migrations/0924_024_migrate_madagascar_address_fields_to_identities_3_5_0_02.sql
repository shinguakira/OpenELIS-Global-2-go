-- source: liquibase liquibase/3.5.x.x/024-migrate-madagascar-address-fields-to-identities.xml::3.5.0-024-drop-migrated-madagascar-person-address-columns::openelisglobal
-- +goose Up
-- +goose StatementBegin
-- Drop Madagascar-specific person address columns after data has been moved to generic address hierarchy identities.
DO $$
            BEGIN
                IF EXISTS (
                    SELECT 1
                    FROM clinlims.person pe
                    LEFT JOIN clinlims.patient p ON p.person_id = pe.id
                    WHERE (
                        pe.province IS NOT NULL AND pe.province != ''
                        AND (p.id IS NULL OR NOT EXISTS (
                            SELECT 1
                            FROM clinlims.patient_identity pi
                            JOIN clinlims.patient_identity_type pit ON pit.id = pi.identity_type_id
                            WHERE pi.patient_id = p.id
                              AND pit.identity_type = 'ADDRESS_HIERARCHY_0'
                              AND pi.identity_data = pe.province
                        ))
                    ) OR (
                        pe.fokontany IS NOT NULL AND pe.fokontany != ''
                        AND (p.id IS NULL OR NOT EXISTS (
                            SELECT 1
                            FROM clinlims.patient_identity pi
                            JOIN clinlims.patient_identity_type pit ON pit.id = pi.identity_type_id
                            WHERE pi.patient_id = p.id
                              AND pit.identity_type = 'ADDRESS_HIERARCHY_3'
                              AND pi.identity_data = pe.fokontany
                        ))
                    ) OR (
                        pe.hamlet_or_lot IS NOT NULL AND pe.hamlet_or_lot != ''
                        AND (p.id IS NULL OR NOT EXISTS (
                            SELECT 1
                            FROM clinlims.patient_identity pi
                            JOIN clinlims.patient_identity_type pit ON pit.id = pi.identity_type_id
                            WHERE pi.patient_id = p.id
                              AND pit.identity_type = 'ADDRESS_HIERARCHY_4'
                              AND pi.identity_data = pe.hamlet_or_lot
                        ))
                    )
                ) THEN
                    RAISE EXCEPTION 'Cannot drop Madagascar person address columns before all values are migrated';
                END IF;
            END $$;
ALTER TABLE clinlims.person DROP COLUMN IF EXISTS province;
ALTER TABLE clinlims.person DROP COLUMN IF EXISTS fokontany;
ALTER TABLE clinlims.person DROP COLUMN IF EXISTS hamlet_or_lot;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/024-migrate-madagascar-address-fields-to-identities.xml::3.5.0-024-drop-migrated-madagascar-person-address-columns::openelisglobal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
