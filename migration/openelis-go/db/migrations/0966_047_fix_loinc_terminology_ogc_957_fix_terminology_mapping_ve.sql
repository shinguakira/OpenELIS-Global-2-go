-- source: liquibase liquibase/3.5.x.x/047-fix-loinc-terminology.xml::OGC-957-fix-terminology-mapping-version::openelis
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.test_terminology_mapping SET last_updated = lastupdated WHERE last_updated IS NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/047-fix-loinc-terminology.xml::OGC-957-fix-terminology-mapping-version::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
