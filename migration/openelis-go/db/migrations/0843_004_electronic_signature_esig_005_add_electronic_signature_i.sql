-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-005-add-electronic-signature-immutability-trigger::openelisdev
-- +goose Up
-- +goose StatementBegin
-- Create trigger to enforce immutability (signatures cannot be modified or deleted)
CREATE OR REPLACE FUNCTION clinlims.prevent_electronic_signature_modification()
            RETURNS TRIGGER AS $BODY$
            BEGIN
                RAISE EXCEPTION 'Electronic signatures are immutable and cannot be modified, deleted, or truncated (21 CFR Part 11)';
                RETURN NULL;
            END;
            $BODY$ LANGUAGE plpgsql;

DO $$ BEGIN
                IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'electronic_signature_immutability_check') THEN
                    CREATE TRIGGER electronic_signature_immutability_check
                    BEFORE UPDATE OR DELETE ON clinlims.electronic_signature
                    FOR EACH ROW EXECUTE FUNCTION clinlims.prevent_electronic_signature_modification();
                END IF;
            END $$;

DROP TRIGGER IF EXISTS electronic_signature_no_truncate ON clinlims.electronic_signature;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/004-electronic-signature.xml::esig-005-add-electronic-signature-immutability-trigger::openelisdev
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
