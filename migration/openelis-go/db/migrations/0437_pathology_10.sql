-- source: liquibase liquibase/2.8.x.x/pathology.xml::10::csteele
-- +goose Up
-- +goose StatementBegin
ALTER TABLE pathology_slide ADD IF NOT EXISTS file_type VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE pathology_slide DROP COLUMN IF EXISTS file_type;
-- +goose StatementEnd
