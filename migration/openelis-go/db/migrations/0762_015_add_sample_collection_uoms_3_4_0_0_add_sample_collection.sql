-- source: liquibase liquibase/3.4.x.x/015-add-sample-collection-uoms.xml::3.4.0.0-add-sample-collection-uoms::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add sample collection units of measure (mL, uL, tubes, slides)
INSERT INTO clinlims.unit_of_measure (id, name, description, uom_type, lastupdated) VALUES (nextval('clinlims.unit_of_measure_seq'), 'mL', 'Milliliters', 'SAMPLE_COLLECTION', NOW()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.unit_of_measure (id, name, description, uom_type, lastupdated) VALUES (nextval('clinlims.unit_of_measure_seq'), 'uL', 'Microliters', 'SAMPLE_COLLECTION', NOW()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.unit_of_measure (id, name, description, uom_type, lastupdated) VALUES (nextval('clinlims.unit_of_measure_seq'), 'tubes', 'Tubes', 'SAMPLE_COLLECTION', NOW()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.unit_of_measure (id, name, description, uom_type, lastupdated) VALUES (nextval('clinlims.unit_of_measure_seq'), 'slides', 'Slides', 'SAMPLE_COLLECTION', NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/015-add-sample-collection-uoms.xml::3.4.0.0-add-sample-collection-uoms::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
