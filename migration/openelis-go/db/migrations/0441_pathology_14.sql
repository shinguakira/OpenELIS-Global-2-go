-- source: liquibase liquibase/2.8.x.x/pathology.xml::14::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- add test_section_id column to the program table
ALTER TABLE clinlims.program ADD IF NOT EXISTS test_section_id INTEGER;
ALTER TABLE program ADD CONSTRAINT test_section_id_program_fk FOREIGN KEY (test_section_id) REFERENCES test_section (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/pathology.xml::14::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
