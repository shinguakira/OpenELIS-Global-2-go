-- source: liquibase liquibase/3.3.x.x/011-optimize-test-normalized-description.xml::test-011-create-normalized-description-trigger::performance-optimization
-- +goose Up
-- +goose StatementBegin
-- Create trigger to automatically update normalized_description when description changes.
--             Ensures data consistency without manual intervention.
CREATE OR REPLACE FUNCTION update_test_normalized_description()
            RETURNS TRIGGER
            LANGUAGE plpgsql
            AS '
            DECLARE
                base_name TEXT;
                sample_type TEXT;
                paren_start INTEGER;
                paren_end INTEGER;
            BEGIN
                IF NEW.description IS NULL THEN
                    NEW.normalized_description = '''';
                    RETURN NEW;
                END IF;

                -- Check for parentheses content (sample type)
                paren_start = POSITION(''('' IN NEW.description);
                paren_end = POSITION('')'' IN NEW.description);

                IF paren_start > 0 AND paren_end > paren_start THEN
                    -- Extract sample type from parentheses
                    sample_type = SUBSTRING(NEW.description FROM paren_start + 1 FOR paren_end - paren_start - 1);
                    sample_type = LOWER(REGEXP_REPLACE(UNACCENT(sample_type), ''[^a-zA-Z0-9]'', '''', ''g''));

                    -- Get base name without parentheses
                    base_name = SUBSTRING(NEW.description FROM 1 FOR paren_start - 1);
                ELSE
                    sample_type = '''';
                    base_name = NEW.description;
                END IF;

                -- Normalize base name
                base_name = LOWER(REGEXP_REPLACE(UNACCENT(base_name), ''[^a-zA-Z0-9]'', '''', ''g''));

                -- Combine base name with sample type
                NEW.normalized_description = base_name || sample_type;

                RETURN NEW;
            END;
            ';

DROP TRIGGER IF EXISTS trigger_update_normalized_description ON test;

CREATE TRIGGER trigger_update_normalized_description
                BEFORE INSERT OR UPDATE OF description ON test
                FOR EACH ROW
                EXECUTE FUNCTION update_test_normalized_description();
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/011-optimize-test-normalized-description.xml::test-011-create-normalized-description-trigger::performance-optimization
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
