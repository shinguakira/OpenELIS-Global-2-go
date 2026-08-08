-- source: liquibase liquibase/3.4.x.x/018-uom-type-mapping-table.xml::3.4.0.0-create-uom-type-map-table::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Create uom_type_map table to map UOMs to usage types
CREATE SEQUENCE  IF NOT EXISTS clinlims.uom_type_map_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS clinlims.uom_type_map (id numeric(10, 0) NOT NULL, uom_id numeric(10, 0) NOT NULL, uom_type VARCHAR(20) NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(), CONSTRAINT uom_type_map_pkey PRIMARY KEY (id));
ALTER TABLE clinlims.uom_type_map ADD CONSTRAINT fk_uom_type_map_uom FOREIGN KEY (uom_id) REFERENCES clinlims.unit_of_measure (id) ON DELETE CASCADE;
ALTER TABLE clinlims.uom_type_map ADD CONSTRAINT uq_uom_type_map UNIQUE (uom_id, uom_type);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/018-uom-type-mapping-table.xml::3.4.0.0-create-uom-type-map-table::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
