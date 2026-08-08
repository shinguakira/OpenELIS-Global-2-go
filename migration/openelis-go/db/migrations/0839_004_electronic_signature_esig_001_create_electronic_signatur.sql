-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-001-create-electronic-signature-table::openelisdev
-- +goose Up
-- +goose StatementBegin
-- Create electronic_signature table for 21 CFR Part 11 compliant signatures
CREATE TABLE IF NOT EXISTS clinlims.electronic_signature (id BIGINT NOT NULL, signer_id BIGINT NOT NULL, signer_name_printed VARCHAR(255) NOT NULL, signature_meaning VARCHAR(30) NOT NULL, signed_at TIMESTAMP WITH TIME ZONE NOT NULL, record_type VARCHAR(100) NOT NULL, record_id BIGINT NOT NULL, rejection_reason TEXT, session_signing_sequence INTEGER NOT NULL, auth_method VARCHAR(20) NOT NULL, client_ip VARCHAR(45), user_agent VARCHAR(500), last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT electronic_signature_pkey PRIMARY KEY (id));
ALTER TABLE clinlims.electronic_signature ALTER COLUMN  last_updated SET DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/004-electronic-signature.xml::esig-001-create-electronic-signature-table::openelisdev
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
