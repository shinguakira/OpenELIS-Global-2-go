-- source: liquibase liquibase/3.3.x.x/barcode_expansion.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- create domain for labels
INSERT INTO clinlims.site_information_domain (id, name, description) VALUES (nextval('clinlims.site_information_domain_seq'), 'labels', 'items that pertain to barcodes/labels') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/barcode_expansion.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
