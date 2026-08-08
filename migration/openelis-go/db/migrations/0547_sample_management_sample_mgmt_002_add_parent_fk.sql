-- source: liquibase liquibase/3.2.x.x/sample-management.xml::sample-mgmt-002-add-parent-fk::dev-team
-- +goose Up
-- +goose StatementBegin
ALTER TABLE sample_item ADD CONSTRAINT fk_sample_item_parent FOREIGN KEY (parent_sample_item_id) REFERENCES sample_item (id) ON UPDATE CASCADE ON DELETE RESTRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sample_item DROP COLUMN IF EXISTS CONSTRAINT;
-- +goose StatementEnd
