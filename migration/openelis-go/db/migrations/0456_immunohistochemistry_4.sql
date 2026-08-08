-- source: liquibase liquibase/2.8.x.x/immunohistochemistry.xml::4::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- add columns to the immunohistochemistry_sample table
ALTER TABLE clinlims.immunohistochemistry_sample ADD IF NOT EXISTS pathology_sample_id INTEGER;
ALTER TABLE clinlims.immunohistochemistry_sample ADD IF NOT EXISTS reffered BOOLEAN;
ALTER TABLE immunohistochemistry_sample ADD CONSTRAINT immunohistochemistry_pathology_sample_id_fk FOREIGN KEY (pathology_sample_id) REFERENCES pathology_sample (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/immunohistochemistry.xml::4::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
