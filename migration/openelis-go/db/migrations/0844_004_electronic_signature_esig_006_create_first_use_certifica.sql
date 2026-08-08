-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-006-create-first-use-certification-table::openelisdev
-- +goose Up
-- +goose StatementBegin
-- Create table for first-use certification per 21 CFR Part 11 §11.100(c)
CREATE TABLE IF NOT EXISTS clinlims.esig_first_use_certification (id BIGINT NOT NULL, user_id BIGINT NOT NULL, certified_at TIMESTAMP WITH TIME ZONE NOT NULL, certification_text TEXT NOT NULL, client_ip VARCHAR(45), user_agent VARCHAR(500), last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT esig_first_use_certification_pkey PRIMARY KEY (id), UNIQUE (user_id));
ALTER TABLE clinlims.esig_first_use_certification ALTER COLUMN  last_updated SET DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/004-electronic-signature.xml::esig-006-create-first-use-certification-table::openelisdev
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
