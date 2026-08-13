-- source: liquibase liquibase/2.7.x.x/add_hpv_testing.xml::23111701::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.localization (id, description, english, french) VALUES (nextval('clinlims.localization_seq'), 'HPV HR', 'HR HPV', 'HPV HR') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_hpv_testing.xml::23111701::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
