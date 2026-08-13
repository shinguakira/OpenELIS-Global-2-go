-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-011-register-first-use-certification-reference-table::openelisdev
-- +goose Up
-- +goose StatementBegin
-- Register esig_first_use_certification in reference_tables for audit trail support
INSERT INTO clinlims.reference_tables (id, name, keep_history, is_hl7_encoded, lastupdated)
            VALUES (
                nextval('clinlims.reference_tables_seq'),
                'ESIG_FIRST_USE_CERTIFICATION',
                'Y',
                'N',
                CURRENT_TIMESTAMP
            ) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/004-electronic-signature.xml::esig-011-register-first-use-certification-reference-table::openelisdev
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
