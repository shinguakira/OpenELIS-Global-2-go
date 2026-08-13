-- source: liquibase liquibase/3.5.x.x/024-migrate-madagascar-address-fields-to-identities.xml::3.5.0-024-migrate-person-hamlet-or-lot-to-address-hierarchy-4::openelisglobal
-- +goose Up
-- +goose StatementBegin
-- Move existing Hamlet/Lot values to generic ADDRESS_HIERARCHY_4 patient identities.
INSERT INTO clinlims.patient_identity (id, identity_type_id, patient_id, identity_data, lastupdated)
            SELECT nextval('clinlims.patient_identity_seq'), pit.id, p.id, pe.hamlet_or_lot, now()
            FROM clinlims.person pe
            JOIN clinlims.patient p ON p.person_id = pe.id
            JOIN clinlims.patient_identity_type pit ON pit.identity_type = 'ADDRESS_HIERARCHY_4'
            WHERE pe.hamlet_or_lot IS NOT NULL
              AND pe.hamlet_or_lot != ''
              AND NOT EXISTS (
                  SELECT 1
                  FROM clinlims.patient_identity existing
                  WHERE existing.patient_id = p.id
                    AND existing.identity_type_id = pit.id
              ) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/024-migrate-madagascar-address-fields-to-identities.xml::3.5.0-024-migrate-person-hamlet-or-lot-to-address-hierarchy-4::openelisglobal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
