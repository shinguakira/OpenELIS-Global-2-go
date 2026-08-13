-- source: liquibase liquibase/3.5.x.x/041-result-components.xml::OGC-937-component-backfill::OGC
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.test_result_component
    (id, test_id, code, label, display_order, result_type, uom_id,
     significant_digits, allow_multiple_readings, is_active, lastupdated)
SELECT gen_random_uuid()::varchar, t.id, 'PRIMARY', t.name, 0,
       (SELECT tr.tst_rslt_type FROM clinlims.test_result tr
         WHERE tr.test_id = t.id AND tr.tst_rslt_type IS NOT NULL
         ORDER BY tr.id LIMIT 1),
       t.uom_id,
       (SELECT tr.significant_digits FROM clinlims.test_result tr
         WHERE tr.test_id = t.id AND tr.significant_digits IS NOT NULL
         ORDER BY tr.id LIMIT 1),
       false, 'Y', now()
  FROM clinlims.test t
 WHERE NOT EXISTS (SELECT 1 FROM clinlims.test_result_component c
                    WHERE c.test_id = t.id AND c.code = 'PRIMARY') ON CONFLICT DO NOTHING;
UPDATE clinlims.result_limits rl
   SET component_id = c.id
  FROM clinlims.test_result_component c
 WHERE c.test_id = rl.test_id AND c.code = 'PRIMARY'
   AND rl.component_id IS NULL;
UPDATE clinlims.test_result tr
   SET component_id = c.id
  FROM clinlims.test_result_component c
 WHERE c.test_id = tr.test_id AND c.code = 'PRIMARY'
   AND tr.component_id IS NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/041-result-components.xml::OGC-937-component-backfill::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
