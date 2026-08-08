-- source: liquibase liquibase/3.4.14.x/004-fix-analyzer-fk-types-and-cascade.xml::014-004-02-fix-notebook-fk-type::openelis
-- +goose Up
-- +goose StatementBegin
-- Fix notebook_analysers.analyser_id type (INTEGER->NUMERIC) and set ON DELETE RESTRICT
ALTER TABLE notebook_analysers DROP CONSTRAINT fk_analysers_analyser;

ALTER TABLE notebook_analysers ALTER COLUMN analyser_id TYPE numeric(10, 0) USING (analyser_id::numeric(10, 0));

ALTER TABLE notebook_analysers ADD CONSTRAINT fk_analysers_analyser FOREIGN KEY (analyser_id) REFERENCES analyzer (id) ON DELETE RESTRICT;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/004-fix-analyzer-fk-types-and-cascade.xml::014-004-02-fix-notebook-fk-type::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
