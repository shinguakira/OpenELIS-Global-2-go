-- source: liquibase liquibase/2.7.x.x/add_electronic_rejection.xml::2403011::CIV developer Group
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.electronic_order ADD IF NOT EXISTS reject_reason VARCHAR(255);
ALTER TABLE clinlims.electronic_order ADD IF NOT EXISTS reject_reason_id VARCHAR(255);
ALTER TABLE clinlims.electronic_order ADD IF NOT EXISTS reject_comment VARCHAR(255);
COMMENT ON COLUMN electronic_order.reject_reason_id IS 'Id from qa_event table';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_electronic_rejection.xml::2403011::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
