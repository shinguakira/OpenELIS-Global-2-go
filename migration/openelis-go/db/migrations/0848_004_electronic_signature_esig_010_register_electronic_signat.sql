-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-010-register-electronic-signature-reference-table::openelisdev
-- +goose Up
-- +goose StatementBegin
-- Register electronic_signature in reference_tables for audit trail support
INSERT INTO clinlims.reference_tables (id, name, keep_history, is_hl7_encoded, lastupdated)
            VALUES (
                nextval('clinlims.reference_tables_seq'),
                'ELECTRONIC_SIGNATURE',
                'Y',
                'N',
                CURRENT_TIMESTAMP
            ) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/004-electronic-signature.xml::esig-010-register-electronic-signature-reference-table::openelisdev
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
