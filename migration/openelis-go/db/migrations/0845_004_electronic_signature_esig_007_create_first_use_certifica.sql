-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-007-create-first-use-certification-sequence::openelisdev
-- +goose Up
-- +goose StatementBegin
-- Create sequence for esig_first_use_certification primary keys
CREATE SEQUENCE  IF NOT EXISTS clinlims.esig_first_use_certification_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS clinlims.esig_first_use_certification_seq;
-- +goose StatementEnd
