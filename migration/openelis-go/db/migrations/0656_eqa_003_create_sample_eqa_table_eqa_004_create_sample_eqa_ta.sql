-- source: liquibase liquibase/3.3.x.x/eqa-003-create-sample-eqa-table.xml::eqa-004-create-sample-eqa-table::eqa-module-feature
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sample_eqa (id numeric(10, 0) NOT NULL, sample_id numeric(10, 0) NOT NULL, is_eqa_sample BOOLEAN DEFAULT FALSE NOT NULL, eqa_program_id numeric(10, 0), eqa_provider_organization_id numeric(10, 0), eqa_provider_sample_id VARCHAR(100), eqa_participant_id VARCHAR(100), eqa_deadline TIMESTAMP WITHOUT TIME ZONE, eqa_priority VARCHAR(20) DEFAULT 'STANDARD', eqa_distribution_id numeric(10, 0), sys_user_id VARCHAR(36) NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT sample_eqa_pkey PRIMARY KEY (id), CONSTRAINT fk_sample_eqa_program FOREIGN KEY (eqa_program_id) REFERENCES eqa_program(id), CONSTRAINT fk_sample_eqa_provider_org FOREIGN KEY (eqa_provider_organization_id) REFERENCES organization(id), CONSTRAINT fk_sample_eqa_sample FOREIGN KEY (sample_id) REFERENCES sample(id), CONSTRAINT fk_sample_eqa_distribution FOREIGN KEY (eqa_distribution_id) REFERENCES eqa_distribution(id), UNIQUE (sample_id));
CREATE INDEX IF NOT EXISTS idx_sample_eqa_sample_id ON sample_eqa(sample_id);
CREATE INDEX IF NOT EXISTS idx_sample_eqa_program_id ON sample_eqa(eqa_program_id);
CREATE INDEX IF NOT EXISTS idx_sample_eqa_deadline ON sample_eqa(eqa_deadline);
CREATE INDEX IF NOT EXISTS idx_sample_eqa_is_eqa ON sample_eqa(is_eqa_sample);
CREATE SEQUENCE  IF NOT EXISTS sample_eqa_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-003-create-sample-eqa-table.xml::eqa-004-create-sample-eqa-table::eqa-module-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
