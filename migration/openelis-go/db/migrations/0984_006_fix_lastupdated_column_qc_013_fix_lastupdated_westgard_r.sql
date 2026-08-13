-- source: liquibase liquibase/qc/006-fix-lastupdated-column.xml::qc-013-fix-lastupdated-westgard-rule-config::openelisglobal
-- +goose Up
-- +goose StatementBegin
ALTER TABLE westgard_rule_config RENAME COLUMN lastupdated TO last_updated;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/qc/006-fix-lastupdated-column.xml::qc-013-fix-lastupdated-westgard-rule-config::openelisglobal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
