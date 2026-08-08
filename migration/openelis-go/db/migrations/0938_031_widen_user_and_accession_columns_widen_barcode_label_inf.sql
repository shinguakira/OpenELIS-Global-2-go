-- source: liquibase liquibase/3.5.x.x/031-widen-user-and-accession-columns.xml::widen-barcode-label-info-code-to-30::openelis
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.barcode_label_info ALTER COLUMN code TYPE VARCHAR(30) USING (code::VARCHAR(30));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/031-widen-user-and-accession-columns.xml::widen-barcode-label-info-code-to-30::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
