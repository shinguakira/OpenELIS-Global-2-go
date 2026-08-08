-- source: liquibase liquibase/2.7.x.x/add_tb_samples.xml::202309156::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.localization (id, description, english, french, lastupdated) VALUES (nextval('clinlims.localization_seq'), 'sampleType name', 'Pericardial puncture fluid', 'Liquide de ponction péricardique', NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_samples.xml::202309156::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
