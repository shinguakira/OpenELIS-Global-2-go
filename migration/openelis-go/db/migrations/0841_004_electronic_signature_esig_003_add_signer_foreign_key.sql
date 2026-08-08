-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-003-add-signer-foreign-key::openelisdev
-- +goose Up
-- +goose StatementBegin
-- Add foreign key from signer_id to system_user to prevent orphaned references
ALTER TABLE clinlims.electronic_signature ADD CONSTRAINT fk_esig_signer_system_user FOREIGN KEY (signer_id) REFERENCES clinlims.system_user (id) ON DELETE RESTRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.electronic_signature DROP COLUMN IF EXISTS CONSTRAINT;
-- +goose StatementEnd
