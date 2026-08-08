-- source: liquibase liquibase/3.5.x.x/042-handling-uom-displayorder.xml::OGC-938-sampletype-test-display-order::OGC
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.sampletype_test ADD IF NOT EXISTS display_order INTEGER;
UPDATE clinlims.sampletype_test st
               SET display_order = sub.rn
              FROM (SELECT id, row_number() OVER (PARTITION BY sample_type_id ORDER BY test_id) AS rn
                      FROM clinlims.sampletype_test) sub
             WHERE st.id = sub.id AND st.display_order IS NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/042-handling-uom-displayorder.xml::OGC-938-sampletype-test-display-order::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
