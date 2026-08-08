-- source: liquibase liquibase/2.7.x.x/add_tb_samples.xml::2023091524::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.lab_unit_role_map (lab_unit_role_map_id, lab_unit) VALUES (nextval('clinlims.lab_unit_role_map_lab_unit_role_map_id_seq'), (SELECT id FROM clinlims.test_section WHERE name ='TB' limit 1)) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_samples.xml::2023091524::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
