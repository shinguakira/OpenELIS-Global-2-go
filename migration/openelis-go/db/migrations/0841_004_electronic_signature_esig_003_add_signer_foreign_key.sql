-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-003-add-signer-foreign-key::openelisdev
-- +goose Up
-- +goose StatementBegin
-- Add foreign key from signer_id to system_user to prevent orphaned references
ALTER TABLE clinlims.electronic_signature ADD CONSTRAINT fk_esig_signer_system_user FOREIGN KEY (signer_id) REFERENCES clinlims.system_user (id) ON DELETE RESTRICT;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/004-electronic-signature.xml::esig-003-add-signer-foreign-key::openelisdev
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
