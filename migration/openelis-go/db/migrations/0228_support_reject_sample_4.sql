-- source: liquibase liquibase/2.7.x.x/support_reject_sample.xml::4::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- Remove 'Please submit another sample. Need to re-test .' From the rejection Reasons
UPDATE clinlims.dictionary SET dictionary_category_id = NULL WHERE dict_entry='Please submit another sample. Need to re-test .';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/support_reject_sample.xml::4::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
