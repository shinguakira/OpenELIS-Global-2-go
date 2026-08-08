-- source: liquibase liquibase/3.3.x.x/029-add-gps-coordinates-to-sample.xml::OGC-300-2026-02-11-1::mherman22
-- +goose Up
-- +goose StatementBegin
ALTER TABLE sample ADD IF NOT EXISTS gps_latitude DECIMAL(9, 6);
ALTER TABLE sample ADD IF NOT EXISTS gps_longitude DECIMAL(9, 6);
ALTER TABLE sample ADD IF NOT EXISTS gps_accuracy_meters INTEGER;
ALTER TABLE sample ADD IF NOT EXISTS gps_capture_method VARCHAR(10);
ALTER TABLE sample ADD IF NOT EXISTS gps_capture_timestamp TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE sample
            ADD CONSTRAINT chk_gps_latitude_range
            CHECK (gps_latitude IS NULL OR (gps_latitude >= -90 AND gps_latitude <= 90));
ALTER TABLE sample
            ADD CONSTRAINT chk_gps_longitude_range
            CHECK (gps_longitude IS NULL OR (gps_longitude >= -180 AND gps_longitude <= 180));
ALTER TABLE sample
            ADD CONSTRAINT chk_gps_accuracy_positive
            CHECK (gps_accuracy_meters IS NULL OR gps_accuracy_meters > 0);
ALTER TABLE sample
            ADD CONSTRAINT chk_gps_capture_method_values
            CHECK (gps_capture_method IS NULL OR gps_capture_method IN ('AUTO', 'MANUAL'));
ALTER TABLE sample
            ADD CONSTRAINT chk_gps_coordinate_consistency
            CHECK ((gps_latitude IS NULL AND gps_longitude IS NULL)
                OR (gps_latitude IS NOT NULL AND gps_longitude IS NOT NULL));
CREATE INDEX IF NOT EXISTS idx_sample_gps_coordinates ON sample(gps_latitude, gps_longitude);
CREATE INDEX IF NOT EXISTS idx_sample_gps_capture_method ON sample(gps_capture_method);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/029-add-gps-coordinates-to-sample.xml::OGC-300-2026-02-11-1::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
