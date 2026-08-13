-- source: liquibase liquibase/2.1.x.x/external_connections.xml::5::csteele
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.basic_authentication_data ADD CONSTRAINT unique_basic_authentication_data_external_connection_id UNIQUE (external_connection_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/external_connections.xml::5::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
