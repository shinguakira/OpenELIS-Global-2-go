-- source: liquibase liquibase/3.3.x.x/009-create-print-history-table.xml::storage-011-create-print-history-table::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS storage_location_print_history (id UUID NOT NULL, location_type VARCHAR(20) NOT NULL, location_id VARCHAR(36) NOT NULL, short_code VARCHAR(10), printed_by VARCHAR(36) NOT NULL, printed_date TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, print_count INTEGER DEFAULT 1 NOT NULL, CONSTRAINT storage_location_print_history_pkey PRIMARY KEY (id));
ALTER TABLE clinlims.storage_location_print_history
            ADD CONSTRAINT chk_location_type
            CHECK (location_type IN ('device', 'shelf', 'rack'));
CREATE INDEX IF NOT EXISTS idx_print_history_location ON storage_location_print_history(location_type, location_id);
CREATE INDEX IF NOT EXISTS idx_print_history_date ON storage_location_print_history(printed_date);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/009-create-print-history-table.xml::storage-011-create-print-history-table::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
