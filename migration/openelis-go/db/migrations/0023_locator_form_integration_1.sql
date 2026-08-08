-- source: liquibase liquibase/2.1.x.x/locator_form_integration.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.status_of_sample (id, description, code, status_type, lastupdated, name, display_key, is_active) VALUES (nextval('clinlims.status_of_sample_seq'), 'This order is non-conforming', '1', 'EXTERNAL_ORDER', NOW(), 'NonConforming', 'status.order.nonConforming', 'Y') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/locator_form_integration.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
