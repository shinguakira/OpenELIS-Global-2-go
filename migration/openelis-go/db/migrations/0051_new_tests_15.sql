-- source: liquibase liquibase/2.3.x.x/new_tests.xml::15::rossumg
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.localization(
                id, description, english, french, lastupdated)
                VALUES
                 (nextval('localization_seq'),'type of sample name','Respiratory Swab','Respiratory Swab', now()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::15::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
