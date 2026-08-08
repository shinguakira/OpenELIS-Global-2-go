-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-013-add-session-timeout-setting::openelisdev
-- +goose Up
-- +goose StatementBegin
-- Add configurable session timeout for e-signature signing sessions (default 30 min per 21 CFR Part 11 industry standard)
INSERT INTO clinlims.site_information (id, name, value, description, value_type, domain_id, lastupdated) VALUES (nextval('clinlims.site_information_seq'), 'esigSessionTimeoutMinutes', '30', 'E-signature session inactivity timeout in minutes (default: 30)', 'text', (SELECT id FROM site_information_domain WHERE name = 'siteIdentity'), NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/004-electronic-signature.xml::esig-013-add-session-timeout-setting::openelisdev
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
