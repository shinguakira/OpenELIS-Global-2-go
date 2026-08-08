-- source: liquibase liquibase/2.8.x.x/cytology.xml::4::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- create cytology_slide table
CREATE SEQUENCE  IF NOT EXISTS clinlims.cytology_slide_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS cytology_slide (id INTEGER NOT NULL, cytology_sample_id INTEGER, slide_number INTEGER, image BYTEA, file_type VARCHAR(255), location VARCHAR(255), last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT cytology_slide_pkey PRIMARY KEY (id));
ALTER TABLE cytology_slide ADD CONSTRAINT cytology_slide_cytology_sample_id_fk FOREIGN KEY (cytology_sample_id) REFERENCES cytology_sample (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/cytology.xml::4::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
