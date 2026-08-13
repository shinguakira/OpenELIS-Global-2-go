-- source: liquibase liquibase/3.5.x.x/047-fix-loinc-terminology.xml::OGC-957-fix-loinc-mapping-relationship::openelis
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.test_terminology_mapping SET relationship = 'SAME_AS' WHERE relationship = 'EQUIVALENT';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/047-fix-loinc-terminology.xml::OGC-957-fix-loinc-mapping-relationship::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
