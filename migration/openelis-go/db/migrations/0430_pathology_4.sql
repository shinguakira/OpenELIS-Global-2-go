-- source: liquibase liquibase/2.8.x.x/pathology.xml::4::csteele
-- +goose Up
-- +goose StatementBegin
-- create pathology_slide table
CREATE SEQUENCE  IF NOT EXISTS clinlims.pathology_slide_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS pathology_slide (id INTEGER NOT NULL, pathology_sample_id INTEGER, slide_number INTEGER, image BYTEA, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pathology_slide_pkey PRIMARY KEY (id));
ALTER TABLE pathology_slide ADD CONSTRAINT pathology_sample_pathology_slide_id_fk FOREIGN KEY (pathology_sample_id) REFERENCES pathology_sample (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/pathology.xml::4::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
