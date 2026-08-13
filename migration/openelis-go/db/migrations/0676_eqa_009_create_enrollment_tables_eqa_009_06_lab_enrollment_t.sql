-- source: liquibase liquibase/3.3.x.x/eqa-009-create-enrollment-tables.xml::eqa-009-06-lab-enrollment-test-map::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Create test/panel mapping table for self-enrollment with CHECK constraint
CREATE TABLE IF NOT EXISTS clinlims.eqa_lab_enrollment_test_map (id numeric(10, 0) NOT NULL, enrollment_id numeric(10, 0) NOT NULL, test_id numeric(10, 0), panel_id numeric(10, 0), sys_user_id VARCHAR(20) NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(), CONSTRAINT eqa_lab_enrollment_test_map_pkey PRIMARY KEY (id), CONSTRAINT fk_test_map_panel FOREIGN KEY (panel_id) REFERENCES clinlims.panel(id), CONSTRAINT fk_test_map_test FOREIGN KEY (test_id) REFERENCES clinlims.test(id), CONSTRAINT fk_test_map_enrollment FOREIGN KEY (enrollment_id) REFERENCES clinlims.eqa_lab_program_enrollment(id));
ALTER TABLE clinlims.eqa_lab_enrollment_test_map
      ADD CONSTRAINT chk_test_or_panel
      CHECK (test_id IS NOT NULL OR panel_id IS NOT NULL);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-009-create-enrollment-tables.xml::eqa-009-06-lab-enrollment-test-map::eqa-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
