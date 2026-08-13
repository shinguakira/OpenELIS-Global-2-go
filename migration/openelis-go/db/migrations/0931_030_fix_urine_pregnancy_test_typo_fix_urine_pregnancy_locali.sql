-- source: liquibase liquibase/3.5.x.x/030-fix-urine-pregnancy-test-typo.xml::fix-urine-pregnancy-localization-value::openelis
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.localization_value SET last_updated = NOW(), value = 'Urine pregnancy test' WHERE locale = 'en' AND value = 'Urine prenancy test';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/030-fix-urine-pregnancy-test-typo.xml::fix-urine-pregnancy-localization-value::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
