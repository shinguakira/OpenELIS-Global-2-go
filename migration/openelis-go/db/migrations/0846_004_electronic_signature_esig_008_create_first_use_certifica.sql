-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-008-create-first-use-certification-index::openelisdev
-- +goose Up
-- +goose StatementBegin
-- Create index for user lookup
CREATE INDEX IF NOT EXISTS idx_esig_cert_user ON clinlims.esig_first_use_certification(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_esig_cert_user;
-- +goose StatementEnd
