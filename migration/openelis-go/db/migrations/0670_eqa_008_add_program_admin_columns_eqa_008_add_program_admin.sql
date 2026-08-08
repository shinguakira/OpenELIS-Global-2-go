-- source: liquibase liquibase/3.3.x.x/eqa-008-add-program-admin-columns.xml::eqa-008-add-program-admin-columns::eqa-module-feature
-- +goose Up
-- +goose StatementBegin
ALTER TABLE eqa_program ADD IF NOT EXISTS organization_id numeric(10, 0);
ALTER TABLE eqa_program ADD CONSTRAINT fk_eqa_program_organization FOREIGN KEY (organization_id) REFERENCES clinlims.organization (id);
ALTER TABLE eqa_program ADD IF NOT EXISTS test_section_id numeric(10, 0);
ALTER TABLE eqa_program ADD CONSTRAINT fk_eqa_program_test_section FOREIGN KEY (test_section_id) REFERENCES clinlims.test_section (id);
ALTER TABLE eqa_program ADD IF NOT EXISTS frequency VARCHAR(50);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-008-add-program-admin-columns.xml::eqa-008-add-program-admin-columns::eqa-module-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
