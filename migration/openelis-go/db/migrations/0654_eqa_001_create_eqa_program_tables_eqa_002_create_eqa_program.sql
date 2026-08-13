-- source: liquibase liquibase/3.3.x.x/eqa-001-create-eqa-program-tables.xml::eqa-002-create-eqa-program-test-table::eqa-module-feature
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS eqa_program_test (id numeric(10, 0) NOT NULL, eqa_program_id numeric(10, 0) NOT NULL, test_id numeric(10, 0) NOT NULL, is_active BOOLEAN DEFAULT TRUE NOT NULL, sys_user_id VARCHAR(36) NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT eqa_program_test_pkey PRIMARY KEY (id), CONSTRAINT fk_eqa_program_test_program FOREIGN KEY (eqa_program_id) REFERENCES eqa_program(id), CONSTRAINT fk_eqa_program_test_test FOREIGN KEY (test_id) REFERENCES test(id));
ALTER TABLE eqa_program_test ADD CONSTRAINT uk_eqa_program_test UNIQUE (eqa_program_id, test_id);
CREATE SEQUENCE  IF NOT EXISTS eqa_program_test_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-001-create-eqa-program-tables.xml::eqa-002-create-eqa-program-test-table::eqa-module-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
