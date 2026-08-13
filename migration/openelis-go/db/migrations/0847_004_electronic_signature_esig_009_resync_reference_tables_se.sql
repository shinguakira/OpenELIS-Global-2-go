-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-009-resync-reference-tables-seq::pmanko
-- +goose Up
-- +goose StatementBegin
-- nce-010 inserted a reference_tables row via COALESCE(MAX(id),0)+1
--             instead of nextval, leaving reference_tables_seq behind.
--             Resync so subsequent nextval calls produce unused IDs.
DO $$
            DECLARE
                max_id BIGINT;
                seq_val BIGINT;
            BEGIN
                SELECT COALESCE(MAX(id), 0) + 1 INTO max_id FROM clinlims.reference_tables;
                SELECT last_value INTO seq_val FROM clinlims.reference_tables_seq;
                IF max_id > seq_val THEN
                    PERFORM setval('clinlims.reference_tables_seq', max_id, false);
                END IF;
            END $$;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/004-electronic-signature.xml::esig-009-resync-reference-tables-seq::pmanko
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
