-- source: liquibase liquibase/3.1.x.x/dictionary-loinc.xml::modify-dictionary-loinc-code::motesamozzy
-- +goose Up
-- +goose StatementBegin
-- Modify LOINC code column length from 10 to 20
ALTER TABLE clinlims.dictionary ALTER COLUMN loinc_code TYPE VARCHAR(20) USING (loinc_code::VARCHAR(20));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.1.x.x/dictionary-loinc.xml::modify-dictionary-loinc-code::motesamozzy
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
