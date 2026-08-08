-- source: liquibase liquibase/3.3.x.x/eqa-006-add-result-audit-columns.xml::eqa-006-add-result-audit-columns::eqa-module-feature
-- +goose Up
-- +goose StatementBegin
ALTER TABLE eqa_result ADD IF NOT EXISTS previous_result_value DECIMAL(15, 5);
ALTER TABLE eqa_result ADD IF NOT EXISTS previous_submission_date TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE eqa_result ADD IF NOT EXISTS previous_submission_method VARCHAR(20);
ALTER TABLE eqa_result ADD IF NOT EXISTS modified_by_user_id numeric(10, 0);
ALTER TABLE eqa_result ADD CONSTRAINT fk_eqa_result_modified_by FOREIGN KEY (modified_by_user_id) REFERENCES system_user (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-006-add-result-audit-columns.xml::eqa-006-add-result-audit-columns::eqa-module-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
