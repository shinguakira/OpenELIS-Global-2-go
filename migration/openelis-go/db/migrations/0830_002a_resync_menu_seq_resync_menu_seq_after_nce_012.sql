-- source: liquibase liquibase/3.5.x.x/002a-resync-menu-seq.xml::resync-menu-seq-after-nce-012::pmanko
-- +goose Up
-- +goose StatementBegin
-- nce-012-add-dashboard-menu inserted a menu row via
--       COALESCE(MAX(id),0)+1 instead of nextval('menu_seq'),
--       leaving the sequence behind the actual max ID. Resync
--       so subsequent nextval calls produce unused IDs.
DO $$
      DECLARE
        max_id BIGINT;
        seq_val BIGINT;
      BEGIN
        SELECT COALESCE(MAX(id), 0) + 1 INTO max_id FROM clinlims.menu;
        SELECT last_value INTO seq_val FROM clinlims.menu_seq;
        -- Only advance the sequence; never move it backward.
        -- On a fresh DB the sequence is behind after nce-012.
        -- On an existing DB where the sequence is already correct, this is a no-op.
        IF max_id > seq_val THEN
          PERFORM setval('clinlims.menu_seq', max_id, false);
        END IF;
      END $$;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/002a-resync-menu-seq.xml::resync-menu-seq-after-nce-012::pmanko
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
