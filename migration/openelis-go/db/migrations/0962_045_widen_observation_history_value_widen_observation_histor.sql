-- source: liquibase liquibase/3.5.x.x/045-widen-observation-history-value.xml::widen-observation-history-value-to-500::openelis
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.observation_history ALTER COLUMN value TYPE VARCHAR(500) USING (value::VARCHAR(500));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/045-widen-observation-history-value.xml::widen-observation-history-value-to-500::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
