-- source: liquibase liquibase/3.3.x.x/034-create-sample-type-request-table.xml::034-create-sample-type-request-table::reagan-meant
-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE  IF NOT EXISTS clinlims.sample_type_request_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS clinlims.sample_type_request (id numeric(10, 0) NOT NULL, sample_id numeric(10, 0) NOT NULL, type_of_sample_id numeric(10, 0) NOT NULL, sort_order numeric(2, 0) DEFAULT 0 NOT NULL, requested_quantity numeric(10, 2) DEFAULT 1, unit_of_measure_id numeric(10, 0), requested_tests TEXT, requested_panels TEXT, status VARCHAR(20) DEFAULT 'REQUESTED' NOT NULL, sample_item_id numeric(10, 0), created_date TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, sysuser_id numeric(10, 0), CONSTRAINT sample_type_request_pkey PRIMARY KEY (id));
ALTER TABLE clinlims.sample_type_request ADD CONSTRAINT fk_sample_type_request_sample FOREIGN KEY (sample_id) REFERENCES clinlims.sample (id) ON DELETE CASCADE;
ALTER TABLE clinlims.sample_type_request ADD CONSTRAINT fk_sample_type_request_type_of_sample FOREIGN KEY (type_of_sample_id) REFERENCES clinlims.type_of_sample (id);
ALTER TABLE clinlims.sample_type_request ADD CONSTRAINT fk_sample_type_request_uom FOREIGN KEY (unit_of_measure_id) REFERENCES clinlims.unit_of_measure (id);
ALTER TABLE clinlims.sample_type_request ADD CONSTRAINT fk_sample_type_request_sample_item FOREIGN KEY (sample_item_id) REFERENCES clinlims.sample_item (id);
CREATE INDEX IF NOT EXISTS idx_sample_type_request_sample_id ON clinlims.sample_type_request(sample_id);
CREATE INDEX IF NOT EXISTS idx_sample_type_request_status ON clinlims.sample_type_request(status);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/034-create-sample-type-request-table.xml::034-create-sample-type-request-table::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
