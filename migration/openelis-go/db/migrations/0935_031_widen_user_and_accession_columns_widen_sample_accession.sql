-- source: liquibase liquibase/3.5.x.x/031-widen-user-and-accession-columns.xml::widen-sample-accession-number-to-25::openelis
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.sample ALTER COLUMN accession_number TYPE VARCHAR(25) USING (accession_number::VARCHAR(25));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/031-widen-user-and-accession-columns.xml::widen-sample-accession-number-to-25::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
