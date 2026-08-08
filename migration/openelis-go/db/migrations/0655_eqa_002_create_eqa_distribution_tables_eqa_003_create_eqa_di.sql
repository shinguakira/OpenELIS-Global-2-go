-- source: liquibase liquibase/3.3.x.x/eqa-002-create-eqa-distribution-tables.xml::eqa-003-create-eqa-distribution-table::eqa-module-feature
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS eqa_distribution (id numeric(10, 0) NOT NULL, fhir_uuid UUID NOT NULL, eqa_program_id numeric(10, 0) NOT NULL, distribution_name VARCHAR(255) NOT NULL, distribution_date TIMESTAMP WITHOUT TIME ZONE NOT NULL, deadline TIMESTAMP WITHOUT TIME ZONE NOT NULL, status VARCHAR(20) DEFAULT 'DRAFT' NOT NULL, created_by numeric(10, 0) NOT NULL, target_value DECIMAL(15, 5), sys_user_id VARCHAR(36) NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT eqa_distribution_pkey PRIMARY KEY (id), CONSTRAINT fk_eqa_dist_created_by FOREIGN KEY (created_by) REFERENCES system_user(id), CONSTRAINT fk_eqa_dist_program FOREIGN KEY (eqa_program_id) REFERENCES eqa_program(id), UNIQUE (fhir_uuid));
CREATE INDEX IF NOT EXISTS idx_eqa_dist_program ON eqa_distribution(eqa_program_id);
CREATE INDEX IF NOT EXISTS idx_eqa_dist_status ON eqa_distribution(status);
CREATE SEQUENCE  IF NOT EXISTS eqa_distribution_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-002-create-eqa-distribution-tables.xml::eqa-003-create-eqa-distribution-table::eqa-module-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
