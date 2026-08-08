-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-004-create-electronic-signature-indexes::openelisdev
-- +goose Up
-- +goose StatementBegin
-- Create indexes for efficient querying of signatures
CREATE INDEX IF NOT EXISTS idx_esig_record ON clinlims.electronic_signature(record_type, record_id);
CREATE INDEX IF NOT EXISTS idx_esig_signer ON clinlims.electronic_signature(signer_id, signed_at);
CREATE INDEX IF NOT EXISTS idx_esig_signed_at ON clinlims.electronic_signature(signed_at);
CREATE INDEX IF NOT EXISTS idx_esig_meaning ON clinlims.electronic_signature(signature_meaning);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/004-electronic-signature.xml::esig-004-create-electronic-signature-indexes::openelisdev
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
