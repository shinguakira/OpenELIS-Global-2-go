-- source: liquibase liquibase/3.3.x.x/barcode_expansion.xml::3::csteele
-- +goose Up
-- +goose StatementBegin
-- rename site information entries with more informative names
UPDATE clinlims.site_information SET name = 'specimenLabelCollectionDate' WHERE name = 'collectionDateCheck';

UPDATE clinlims.site_information SET name = 'specimenLabelPatientSex' WHERE name = 'patientSexCheck';

UPDATE clinlims.site_information SET name = 'specimenLabelCollectedBy' WHERE name = 'collectedByCheck';

UPDATE clinlims.site_information SET name = 'specimenLabelTests' WHERE name = 'testsCheck';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/barcode_expansion.xml::3::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
