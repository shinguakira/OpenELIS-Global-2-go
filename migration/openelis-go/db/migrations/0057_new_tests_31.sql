-- source: liquibase liquibase/2.3.x.x/new_tests.xml::31::rossumg
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.type_of_sample SET local_abbrev = 'Resp Swab' WHERE description = 'Respiratory Swab';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::31::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
