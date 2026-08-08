-- source: liquibase liquibase/3.5.x.x/050-tsh-history-fk-cascade.xml::OGC-766-tsh-history-fk-cascade::OGC
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.test_sample_handling_history DROP CONSTRAINT fk_tsh_history_handling;

ALTER TABLE clinlims.test_sample_handling_history ADD CONSTRAINT fk_tsh_history_handling FOREIGN KEY (test_sample_handling_id) REFERENCES clinlims.test_sample_handling (id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/050-tsh-history-fk-cascade.xml::OGC-766-tsh-history-fk-cascade::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
