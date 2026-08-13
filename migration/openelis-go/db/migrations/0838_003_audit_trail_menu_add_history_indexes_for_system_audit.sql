-- source: liquibase liquibase/3.5.x.x/003-audit-trail-menu.xml::add-history-indexes-for-system-audit::systemLevelAudit
-- +goose Up
-- +goose StatementBegin
-- Add indexes to support system audit event queries (filter/sort by timestamp, reference_table, sys_user_id)
CREATE INDEX IF NOT EXISTS idx_history_timestamp_reftable
        ON clinlims.history (timestamp DESC, reference_table);

CREATE INDEX IF NOT EXISTS idx_history_sys_user_id
        ON clinlims.history (sys_user_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/003-audit-trail-menu.xml::add-history-indexes-for-system-audit::systemLevelAudit
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
