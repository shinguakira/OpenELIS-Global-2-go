-- source: liquibase liquibase/3.4.x.x/002-modify-analyzer-table.xml::011-002-02-widen-name-and-add-structural-columns::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Add analyzer_type_id FK to analyzer (has_setup_page and name widening already handled by 2.3.x.x)
ALTER TABLE analyzer ADD IF NOT EXISTS analyzer_type_id numeric(10, 0);
ALTER TABLE analyzer ADD CONSTRAINT fk_analyzer_analyzer_type FOREIGN KEY (analyzer_type_id) REFERENCES analyzer_type (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/002-modify-analyzer-table.xml::011-002-02-widen-name-and-add-structural-columns::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
