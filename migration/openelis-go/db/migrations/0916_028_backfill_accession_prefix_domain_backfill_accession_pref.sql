-- source: liquibase liquibase/3.5.x.x/028-backfill-accession-prefix-domain.xml::backfill-accession-prefix-domain::openelis
-- +goose Up
-- +goose StatementBegin
-- Re-associates the 'Accession number prefix' site_information row with the siteIdentity domain so it appears under Admin > General Configuration > Site Information. The dev-init siteInfo.sql INSERT omitted domain_id, leaving the row orphaned; properties-file round-trip still worked because that path queries by name, but the admin menu (inner-join on domain) filtered the row out.
UPDATE clinlims.site_information SET domain_id = (SELECT id FROM clinlims.site_information_domain WHERE name = 'siteIdentity'), lastupdated = NOW() WHERE name = 'Accession number prefix' AND domain_id IS NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/028-backfill-accession-prefix-domain.xml::backfill-accession-prefix-domain::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
