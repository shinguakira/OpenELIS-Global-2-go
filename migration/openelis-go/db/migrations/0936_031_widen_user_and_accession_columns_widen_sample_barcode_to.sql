-- source: liquibase liquibase/3.5.x.x/031-widen-user-and-accession-columns.xml::widen-sample-barcode-to-25::openelis
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.sample ALTER COLUMN barcode TYPE VARCHAR(25) USING (barcode::VARCHAR(25));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/031-widen-user-and-accession-columns.xml::widen-sample-barcode-to-25::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
