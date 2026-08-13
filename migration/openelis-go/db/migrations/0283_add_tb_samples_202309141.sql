-- source: liquibase liquibase/2.7.x.x/add_tb_samples.xml::202309141::CIV developer Group
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.localization SET french = 'Crachat' WHERE french='Sputum';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_samples.xml::202309141::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
