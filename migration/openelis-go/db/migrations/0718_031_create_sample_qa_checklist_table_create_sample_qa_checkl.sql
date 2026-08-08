-- source: liquibase liquibase/3.3.x.x/031-create-sample-qa-checklist-table.xml::create-sample-qa-checklist-table::reagan-meant
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sample_qa_checklist (id numeric(10, 0) NOT NULL, sample_id numeric(10, 0) NOT NULL, verified_items TEXT, all_required_verified BOOLEAN DEFAULT FALSE NOT NULL, verified_by_user_id numeric(10, 0), verified_date TIMESTAMP WITHOUT TIME ZONE, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT sample_qa_checklist_pkey PRIMARY KEY (id), CONSTRAINT fk_qa_checklist_user FOREIGN KEY (verified_by_user_id) REFERENCES system_user(id), CONSTRAINT fk_qa_checklist_sample FOREIGN KEY (sample_id) REFERENCES sample(id) ON DELETE CASCADE, UNIQUE (sample_id));
CREATE SEQUENCE  IF NOT EXISTS sample_qa_checklist_seq START WITH 1 INCREMENT BY 1;
CREATE INDEX IF NOT EXISTS idx_qa_checklist_sample_id ON sample_qa_checklist(sample_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/031-create-sample-qa-checklist-table.xml::create-sample-qa-checklist-table::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
