-- source: liquibase liquibase/2.6.x.x/method.xml::1::cliff
-- +goose Up
-- +goose StatementBegin
DELETE FROM clinlims.method WHERE is_active = 'Y';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/method.xml::1::cliff
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
