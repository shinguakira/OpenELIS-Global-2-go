-- source: liquibase liquibase/2.1.x.x/barcode_label_info_alter.xml::1::rossumg
-- +goose Up
-- +goose StatementBegin
-- increase barcode_label_info.code length to handle code.x
ALTER TABLE clinlims.barcode_label_info ALTER COLUMN code TYPE VARCHAR(25) USING (code::VARCHAR(25));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/barcode_label_info_alter.xml::1::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
