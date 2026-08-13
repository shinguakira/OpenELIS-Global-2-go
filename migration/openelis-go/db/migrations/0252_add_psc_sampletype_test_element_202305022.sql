-- source: liquibase liquibase/2.7.x.x/add_psc_sampletype_test_element.xml::202305022::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.type_of_sample (id, description, domain, local_abbrev, is_active, name_localization_id, display_key) VALUES (nextval('clinlims.type_of_sample_seq'), 'PSC', 'H', 'PSC', TRUE, (SELECT id FROM clinlims.localization WHERE description = 'PSC'), 'sample.type.PSC') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_psc_sampletype_test_element.xml::202305022::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
