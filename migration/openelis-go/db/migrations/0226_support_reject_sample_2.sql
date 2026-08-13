-- source: liquibase liquibase/2.7.x.x/support_reject_sample.xml::2::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- Add sample rejected Analysis status
INSERT INTO clinlims.status_of_sample (id, code, lastupdated, status_type, name, description, display_key) VALUES ( nextval( 'status_of_sample_seq' ) , '1', NOW(), 'ANALYSIS', 'Sample Rejected', 'The sample has been rejected', 'status.sample.rejected') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.status_of_sample (id, code, lastupdated, status_type, name, description, display_key) VALUES ( nextval( 'status_of_sample_seq' ) , '1', NOW(), 'SAMPLE', 'Sample Rejected', 'The sample has been rejected', 'status.sample.rejected') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/support_reject_sample.xml::2::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
