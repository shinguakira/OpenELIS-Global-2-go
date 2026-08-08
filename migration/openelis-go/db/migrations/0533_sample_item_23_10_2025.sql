-- source: liquibase liquibase/3.2.x.x/sample_item.xml::23-10-2025::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Modify external_id column length from 20 to 40
ALTER TABLE clinlims.sample_item ALTER COLUMN external_id TYPE VARCHAR(40) USING (external_id::VARCHAR(40));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/sample_item.xml::23-10-2025::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
