-- source: liquibase liquibase/2.7.x.x/fix_person_table.xml::1::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Increase size of the person.state column
ALTER TABLE person ALTER COLUMN state TYPE VARCHAR(225) USING (state::VARCHAR(225));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/fix_person_table.xml::1::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
