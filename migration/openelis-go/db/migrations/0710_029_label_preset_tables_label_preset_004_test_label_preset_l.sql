-- source: liquibase liquibase/3.3.x.x/029-label-preset-tables.xml::label-preset-004-test-label-preset-link::ogc-285
-- +goose Up
-- +goose StatementBegin
-- Create test_label_preset_link table: per-test preset links with per-sample qty overrides (FRS §3.5)
CREATE SEQUENCE  IF NOT EXISTS clinlims.test_label_preset_link_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS clinlims.test_label_preset_link (id INTEGER DEFAULT nextval('test_label_preset_link_seq') NOT NULL, test_id numeric(10, 0) NOT NULL, preset_id INTEGER NOT NULL, default_qty INTEGER NOT NULL, max_qty INTEGER NOT NULL, allow_override BOOLEAN DEFAULT TRUE NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT test_label_preset_link_pkey PRIMARY KEY (id), CONSTRAINT fk_test_label_preset_link_preset FOREIGN KEY (preset_id) REFERENCES label_preset(id), CONSTRAINT fk_test_label_preset_link_test FOREIGN KEY (test_id) REFERENCES test(id) ON DELETE CASCADE);
ALTER TABLE clinlims.test_label_preset_link ADD CONSTRAINT test_label_preset_link_uniq UNIQUE (test_id, preset_id);
ALTER TABLE clinlims.test_label_preset_link
             ADD CONSTRAINT test_label_preset_link_qty_nonneg CHECK (default_qty >= 0),
             ADD CONSTRAINT test_label_preset_link_max_gte_default CHECK (max_qty >= default_qty);
CREATE INDEX IF NOT EXISTS idx_test_label_preset_link_test ON clinlims.test_label_preset_link(test_id);
CREATE INDEX IF NOT EXISTS idx_test_label_preset_link_preset ON clinlims.test_label_preset_link(preset_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/029-label-preset-tables.xml::label-preset-004-test-label-preset-link::ogc-285
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
