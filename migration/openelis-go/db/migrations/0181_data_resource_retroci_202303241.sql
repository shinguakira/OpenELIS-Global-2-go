-- source: liquibase liquibase/2.6.x.x/data_resource_retroci.xml::202303241::CIV developer Group
-- +goose Up
-- +goose StatementBegin
-- create data_resource table and map tables in RetroCI database
CREATE TABLE IF NOT EXISTS clinlims.data_resource (
            "id" NUMERIC (20),
            "name" VARCHAR (20),
            "collection_name" VARCHAR (20),
            "header_key" VARCHAR (40),
            "level" VARCHAR (20),
            "indicator_id" NUMERIC (20),
            "lastupdated" timestamp without time zone NOT NULL,
            PRIMARY KEY ("id")
            );
CREATE SEQUENCE IF NOT EXISTS clinlims.data_resource_seq START 1;
INSERT INTO clinlims.reference_tables (id, name, keep_history,
            is_hl7_encoded)
            VALUES (nextval('clinlims.reference_tables_seq'), 'DATA_RESOURCE', 'Y',
            'N') ON CONFLICT DO NOTHING;
CREATE TABLE IF NOT EXISTS clinlims.data_resource_level_id (
            "id" SERIAL,
            "level" VARCHAR (20),
            "id_for_level" VARCHAR (20),
            "data_resource_id" NUMERIC (20),
            PRIMARY KEY ("id")
            );
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/data_resource_retroci.xml::202303241::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
