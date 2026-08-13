-- source: liquibase liquibase/3.5.x.x/042-handling-uom-displayorder.xml::OGC-938-unit-of-measure-master::OGC
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.unit_of_measure ADD IF NOT EXISTS code VARCHAR(20);
ALTER TABLE clinlims.unit_of_measure ADD CONSTRAINT uq_unit_of_measure_code UNIQUE (code);
ALTER TABLE clinlims.unit_of_measure ADD IF NOT EXISTS ucum_code VARCHAR(40);
ALTER TABLE clinlims.unit_of_measure ADD IF NOT EXISTS is_active VARCHAR(2) DEFAULT 'Y' NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/042-handling-uom-displayorder.xml::OGC-938-unit-of-measure-master::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
