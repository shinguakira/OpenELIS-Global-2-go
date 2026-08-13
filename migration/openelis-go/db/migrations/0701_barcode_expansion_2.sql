-- source: liquibase liquibase/3.3.x.x/barcode_expansion.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
-- update existing site information that should point to the labels domain
UPDATE clinlims.site_information SET domain_id = (SELECT id FROM clinlims.site_information_domain WHERE name = 'labels') WHERE name = 'numMaxOrderLabels'
                OR name = 'numMaxSpecimenLabels'
                OR name = 'numDefaultSpecimenLabels'
                OR name = 'numMaxAliquotLabels'
                OR name = 'numDefaultOrderLabels'
                OR name = 'numDefaultAliquotLabels'
                OR name = 'heightOrderLabels'
                OR name = 'widthOrderLabels'
                OR name = 'heightSpecimenLabels'
                OR name = 'widthSpecimenLabels'
                OR name = 'collectionDateCheck'
                OR name = 'patientSexCheck'
                OR name = 'collectedByCheck'
                OR name = 'testsCheck'
                OR name = 'heightBlockLabels'
                OR name = 'widthBlockLabels'
                OR name = 'heightSlideLabels'
                OR name = 'widthSlideLabels'
                OR name = 'heightStorageLocationLabels'
                OR name = 'widthStorageLocationLabels';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/barcode_expansion.xml::2::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
