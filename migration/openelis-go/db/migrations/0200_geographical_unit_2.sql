-- source: liquibase liquibase/2.6.x.x/geographical_unit.xml::2::cliff
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'restrictFreeTextMethodEntry', NOW(), 'Users cannot enter new methods through result entry', 'false', 'false', (SELECT id FROM site_information_domain WHERE name = 'resultConfiguration'), 'boolean', 'instructions.method.limit', '0', 'siteInfo.restrictFreeTextMethodEntry') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/geographical_unit.xml::2::cliff
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
