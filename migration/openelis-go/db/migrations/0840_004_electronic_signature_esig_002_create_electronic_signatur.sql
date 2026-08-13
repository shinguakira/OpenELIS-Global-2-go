-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-002-create-electronic-signature-sequence::openelisdev
-- +goose Up
-- +goose StatementBegin
-- Create sequence for electronic_signature primary keys
CREATE SEQUENCE  IF NOT EXISTS clinlims.electronic_signature_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS clinlims.electronic_signature_seq;
-- +goose StatementEnd
