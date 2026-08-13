-- source: liquibase liquibase/2.8.x.x/sample_types.xml::1.1::csteele
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.type_of_sample SET lastupdated = NOW(), local_abbrev = 'TMP' WHERE description='Tissue post mortem';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/sample_types.xml::1.1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
