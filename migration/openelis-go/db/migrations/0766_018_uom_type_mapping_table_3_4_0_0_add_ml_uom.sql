-- source: liquibase liquibase/3.4.x.x/018-uom-type-mapping-table.xml::3.4.0.0-add-ml-uom::reagan-meant
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.unit_of_measure (id, name, description, lastupdated) VALUES (nextval('clinlims.unit_of_measure_seq'), 'mL', 'Milliliters', NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/018-uom-type-mapping-table.xml::3.4.0.0-add-ml-uom::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
