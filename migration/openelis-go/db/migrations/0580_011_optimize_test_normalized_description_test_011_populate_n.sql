-- source: liquibase liquibase/3.3.x.x/011-optimize-test-normalized-description.xml::test-011-populate-normalized-description::performance-optimization
-- +goose Up
-- +goose StatementBegin
-- Populate normalized_description column with normalized values.
--             Normalization logic: remove parentheses content and non-alphanumeric characters.
UPDATE test
            SET normalized_description = (
                CASE
                    WHEN POSITION('(' IN description) > 0 AND POSITION(')' IN description) > POSITION('(' IN description) THEN
                        -- Has sample type in parentheses
                        LOWER(REGEXP_REPLACE(
                            UNACCENT(SUBSTRING(description FROM 1 FOR POSITION('(' IN description) - 1)),
                            '[^a-zA-Z0-9]', '', 'g'
                        )) ||
                        LOWER(REGEXP_REPLACE(
                            UNACCENT(SUBSTRING(description FROM POSITION('(' IN description) + 1 FOR POSITION(')' IN description) - POSITION('(' IN description) - 1)),
                            '[^a-zA-Z0-9]', '', 'g'
                        ))
                    ELSE
                        -- No sample type in parentheses
                        LOWER(REGEXP_REPLACE(UNACCENT(description), '[^a-zA-Z0-9]', '', 'g'))
                END
            )
            WHERE description IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/011-optimize-test-normalized-description.xml::test-011-populate-normalized-description::performance-optimization
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
