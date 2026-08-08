-- source: liquibase liquibase/3.5.x.x/048-test-reagent-link.xml::OGC-986-test-reagent-link-table::OGC
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.test_reagent_link (id VARCHAR(36) NOT NULL, test_id numeric(10) NOT NULL, reagent_id BIGINT NOT NULL, usage_type VARCHAR(20) NOT NULL, quantity_per_test numeric(15, 6), quantity_unit VARCHAR(50), lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pk_test_reagent_link PRIMARY KEY (id));
ALTER TABLE clinlims.test_reagent_link
            ADD CONSTRAINT test_reagent_link_usage_type_chk
            CHECK (usage_type IN ('PRIMARY', 'SECONDARY'));
ALTER TABLE clinlims.test_reagent_link ADD CONSTRAINT uq_test_reagent_link_test_reagent UNIQUE (test_id, reagent_id);
ALTER TABLE clinlims.test_reagent_link ADD CONSTRAINT fk_test_reagent_link_test FOREIGN KEY (test_id) REFERENCES clinlims.test (id);
ALTER TABLE clinlims.test_reagent_link ADD CONSTRAINT fk_test_reagent_link_inventory_item FOREIGN KEY (reagent_id) REFERENCES clinlims.inventory_item (id);
CREATE INDEX IF NOT EXISTS idx_test_reagent_link_test_usage ON clinlims.test_reagent_link(test_id, usage_type);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/048-test-reagent-link.xml::OGC-986-test-reagent-link-table::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
