-- source: liquibase liquibase/2.7.x.x/add_tb_samples.xml::2023091511::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.type_of_sample (id, description, domain, local_abbrev, is_active, name_localization_id, display_key) VALUES (nextval('clinlims.type_of_sample_seq'), 'Sputum', 'H', 'Sputum', TRUE, (SELECT id FROM clinlims.localization WHERE french = 'Sputum' OR french ='Crachat' limit 1), 'Sample.type.Sputum') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_samples.xml::2023091511::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
