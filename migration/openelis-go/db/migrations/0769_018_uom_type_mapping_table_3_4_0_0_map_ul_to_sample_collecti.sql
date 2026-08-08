-- source: liquibase liquibase/3.4.x.x/018-uom-type-mapping-table.xml::3.4.0.0-map-ul-to-sample-collection::reagan-meant
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.uom_type_map (id, uom_id, uom_type, lastupdated)
      SELECT nextval('clinlims.uom_type_map_seq'), id, 'SAMPLE_COLLECTION', NOW()
      FROM clinlims.unit_of_measure
      WHERE name = 'uL' ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/018-uom-type-mapping-table.xml::3.4.0.0-map-ul-to-sample-collection::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
