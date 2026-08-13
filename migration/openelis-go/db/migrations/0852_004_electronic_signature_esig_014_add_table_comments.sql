-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-014-add-table-comments::openelisdev
-- +goose Up
-- +goose StatementBegin
-- Add documentation comments to tables and columns
COMMENT ON TABLE clinlims.electronic_signature IS
            '21 CFR Part 11 compliant electronic signatures.
            Each record represents a legally-binding signature event.
            Table is append-only (enforced by trigger) - records cannot be modified or deleted.';

COMMENT ON COLUMN clinlims.electronic_signature.signer_name_printed IS
            'Full name of signer at time of signing. Denormalized so name changes do not alter historical records.';

COMMENT ON COLUMN clinlims.electronic_signature.signature_meaning IS
            'Legal meaning per §11.50. Values: AUTHORED (data entry), VALIDATED_AND_RELEASED (approval), REJECTED (rejection).';

COMMENT ON COLUMN clinlims.electronic_signature.session_signing_sequence IS
            'Position in signing session. 1 = full auth (user ID + password), 2+ = password-only per §11.200(a)(1)(i).';

COMMENT ON COLUMN clinlims.electronic_signature.auth_method IS
            'Authentication provider used. Values: LOCAL (OpenELIS database), KEYCLOAK (SSO).';

COMMENT ON TABLE clinlims.esig_first_use_certification IS
            'First-use certification per 21 CFR Part 11 §11.100(c).
            Users must certify that their e-signature is legally binding before first use.
            One record per user. Admin can revoke (delete) to force re-certification.';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/004-electronic-signature.xml::esig-014-add-table-comments::openelisdev
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
