-- source: liquibase liquibase/3.5.x.x/030-fix-urine-pregnancy-test-typo.xml::fix-urine-pregnancy-test-name::openelis
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.test SET lastupdated = NOW(), name = 'Urine pregnancy test-Urine' WHERE name = 'Urine prenancy test-Urine';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/030-fix-urine-pregnancy-test-typo.xml::fix-urine-pregnancy-test-name::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
