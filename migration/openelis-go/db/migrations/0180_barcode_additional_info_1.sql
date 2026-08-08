-- source: liquibase liquibase/2.5.x.x/barcode_additional_info.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- specify max vs deafult for barcode site info
UPDATE clinlims.site_information SET name = 'numMaxOrderLabels' WHERE name = 'numOrderLabels';
UPDATE clinlims.site_information SET name = 'numMaxSpecimenLabels' WHERE name = 'numSpecimenLabels';
UPDATE clinlims.site_information SET name = 'numMaxAliquotLabels' WHERE name = 'numAliquotLabels';
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, value_type, "group") VALUES (nextval('clinlims.site_information_seq'), 'numDefaultOrderLabels', NOW(), 'default number of order labels printed', '2', 'false', 'text', '0') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, value_type, "group") VALUES (nextval('clinlims.site_information_seq'), 'numDefaultSpecimenLabels', NOW(), 'default number of specimen labels printed', '1', 'false', 'text', '0') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, value_type, "group") VALUES (nextval('clinlims.site_information_seq'), 'numDefaultAliquotLabels', NOW(), 'default number of aliquot labels printed', '1', 'false', 'text', '0') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.5.x.x/barcode_additional_info.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
