-- source: liquibase liquibase/3.5.x.x/046-fix-component-version.xml::OGC-937-fix-test-result-component-version::openelis
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.test_result_component SET last_updated = lastupdated WHERE last_updated IS NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/046-fix-component-version.xml::OGC-937-fix-test-result-component-version::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
