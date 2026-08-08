-- source: liquibase liquibase/3.5.x.x/004-analysis-analyzer-id.xml::3.5.0-003-add-analysis-analyzer-id-fk::pmanko
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.analysis ADD CONSTRAINT fk_analysis_analyzer FOREIGN KEY (analyzer_id) REFERENCES clinlims.analyzer (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.analysis DROP COLUMN IF EXISTS CONSTRAINT;
-- +goose StatementEnd
