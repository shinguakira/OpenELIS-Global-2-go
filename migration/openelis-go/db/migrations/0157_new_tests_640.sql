-- source: liquibase liquibase/2.3.x.x/new_tests.xml::640::csteele
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.test t
                SET default_test_result_id = (
                    SELECT tr.id
                    FROM clinlims.test_result tr
                    WHERE CAST (tr.value as NUMERIC) in (
                        SELECT d.id FROM clinlims.dictionary d
                        WHERE d.dict_entry = 'IgM NOT DETECTED'
                    )
                    AND tr.test_id = t.id
                )
                WHERE description = 'COVID-19ANTIBODYIgM(Plasma)';

UPDATE clinlims.result_limits rl
                SET normal_dictionary_id = (
                    SELECT d.id FROM clinlims.dictionary d
                    WHERE d.dict_entry = 'IgM NOT DETECTED'
                )
                WHERE rl.test_id in (
                SELECT id FROM clinlims.test
                WHERE description = 'COVID-19ANTIBODYIgM(Plasma)'
                );
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::640::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
