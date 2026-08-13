-- source: liquibase liquibase/3.5.x.x/031-widen-user-and-accession-columns.xml::widen-login-user-login-name-to-30::openelis
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.login_user ALTER COLUMN login_name TYPE VARCHAR(30) USING (login_name::VARCHAR(30));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/031-widen-user-and-accession-columns.xml::widen-login-user-login-name-to-30::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
