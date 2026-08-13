-- source: liquibase liquibase/3.3.x.x/eqa-004-create-eqa-result-table.xml::eqa-005-create-eqa-result-table::eqa-module-feature
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS eqa_result (id numeric(10, 0) NOT NULL, fhir_uuid UUID NOT NULL, eqa_distribution_id numeric(10, 0) NOT NULL, participant_organization_id numeric(10, 0) NOT NULL, test_id numeric(10, 0) NOT NULL, result_value DECIMAL(15, 5), target_value DECIMAL(15, 5), z_score DECIMAL(10, 5), submission_method VARCHAR(20) NOT NULL, submission_date TIMESTAMP WITHOUT TIME ZONE NOT NULL, performance_status VARCHAR(20), is_late_submission BOOLEAN DEFAULT FALSE NOT NULL, late_submission_justification TEXT, approved_by numeric(10, 0), sys_user_id VARCHAR(36) NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT eqa_result_pkey PRIMARY KEY (id), CONSTRAINT fk_eqa_result_org FOREIGN KEY (participant_organization_id) REFERENCES organization(id), CONSTRAINT fk_eqa_result_distribution FOREIGN KEY (eqa_distribution_id) REFERENCES eqa_distribution(id), CONSTRAINT fk_eqa_result_test FOREIGN KEY (test_id) REFERENCES test(id), CONSTRAINT fk_eqa_result_approved_by FOREIGN KEY (approved_by) REFERENCES system_user(id), UNIQUE (fhir_uuid));
ALTER TABLE eqa_result ADD CONSTRAINT uk_eqa_result_dist_org_test UNIQUE (eqa_distribution_id, participant_organization_id, test_id);
CREATE INDEX IF NOT EXISTS idx_eqa_result_distribution ON eqa_result(eqa_distribution_id);
CREATE INDEX IF NOT EXISTS idx_eqa_result_participant ON eqa_result(participant_organization_id);
CREATE SEQUENCE  IF NOT EXISTS eqa_result_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-004-create-eqa-result-table.xml::eqa-005-create-eqa-result-table::eqa-module-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
