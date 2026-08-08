-- source: liquibase liquibase/2.0.x.x/convert_id_types.xml::2::caleb
-- +goose Up
-- +goose StatementBegin
-- modify gender as a second example
ALTER TABLE clinlims.gender ALTER COLUMN id TYPE INTEGER USING (id::INTEGER);

ALTER TABLE clinlims.gender RENAME COLUMN lastupdated TO last_updated;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.0.x.x/convert_id_types.xml::2::caleb
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
