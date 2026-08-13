-- source: liquibase liquibase/qc/006-fix-lastupdated-column.xml::qc-016-fix-lastupdated-qc-alert::openelisglobal
-- +goose Up
-- +goose StatementBegin
ALTER TABLE qc_alert RENAME COLUMN lastupdated TO last_updated;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/qc/006-fix-lastupdated-column.xml::qc-016-fix-lastupdated-qc-alert::openelisglobal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
