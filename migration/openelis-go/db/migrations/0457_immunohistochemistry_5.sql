-- source: liquibase liquibase/2.8.x.x/immunohistochemistry.xml::5::csteele
-- +goose Up
-- +goose StatementBegin
-- add immunohistochemistry sample type info
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'sampleType name', 'Immunohistochemistry specimen', 'Immunohistochemistry specimen') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.type_of_sample (id, description, lastupdated, domain, local_abbrev, display_key, name_localization_id) VALUES ( nextval( 'type_of_sample_seq' ) , 'Immunohistochemistry specimen', NOW(), 'H', 'IMMUNO', 'sample.type.immunohistochemistrySpecimen', (select id from localization where english = 'Immunohistochemistry specimen' and description = 'sampleType name' limit 1)) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/immunohistochemistry.xml::5::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
