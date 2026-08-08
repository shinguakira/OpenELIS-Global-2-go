-- source: liquibase liquibase/2.8.x.x/cytology.xml::2::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- create cytology_specimen_adequacy_value table
CREATE TABLE IF NOT EXISTS cytology_specimen_adequacy_value (value VARCHAR(255), cytology_specimen_adequacy_id INTEGER);
ALTER TABLE cytology_specimen_adequacy_value ADD CONSTRAINT cytology_specimen_adequacy_value_cytology_specimen_adequacy_id_fk FOREIGN KEY (cytology_specimen_adequacy_id) REFERENCES cytology_specimen_adequacy (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/cytology.xml::2::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
