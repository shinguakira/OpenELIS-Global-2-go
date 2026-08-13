-- source: liquibase liquibase/2.7.x.x/support_reject_sample.xml::1::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- add rejected column to sample_item table
ALTER TABLE clinlims.sample_item ADD IF NOT EXISTS rejected BOOLEAN DEFAULT FALSE;
ALTER TABLE clinlims.sample_item ADD IF NOT EXISTS reject_reason_id numeric(10);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/support_reject_sample.xml::1::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
