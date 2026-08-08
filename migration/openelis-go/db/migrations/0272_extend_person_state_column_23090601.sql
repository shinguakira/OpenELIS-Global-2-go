-- source: liquibase liquibase/2.7.x.x/extend_person_state_column.xml::23090601::CIV Developer Group
-- +goose Up
-- +goose StatementBegin
-- Extend patient state length
ALTER TABLE clinlims.person ALTER COLUMN state TYPE VARCHAR(100) USING (state::VARCHAR(100));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/extend_person_state_column.xml::23090601::CIV Developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
