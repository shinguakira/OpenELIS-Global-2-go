-- source: liquibase liquibase/2.7.x.x/add_tb_tests.xml::2023092201::CIV developer Group
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.tb_method_test ALTER COLUMN method_id TYPE	numeric(10) USING method_id::numeric;

ALTER TABLE clinlims.tb_method_test ALTER COLUMN test_id TYPE numeric(10) USING test_id::numeric;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_tests.xml::2023092201::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
