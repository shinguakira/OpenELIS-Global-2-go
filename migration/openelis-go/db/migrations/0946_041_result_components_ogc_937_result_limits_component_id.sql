-- source: liquibase liquibase/3.5.x.x/041-result-components.xml::OGC-937-result-limits-component-id::OGC
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.result_limits ADD IF NOT EXISTS component_id VARCHAR(36);
ALTER TABLE clinlims.result_limits ADD CONSTRAINT fk_result_limits_component FOREIGN KEY (component_id) REFERENCES clinlims.test_result_component (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/041-result-components.xml::OGC-937-result-limits-component-id::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
