-- source: liquibase liquibase/3.5.x.x/shipment-012-migrate-to-sample-item.xml::2::pkomena
-- +goose Up
-- +goose StatementBegin
-- Migrate data from box_sample to box_sample_item - create entry for each SampleItem of the Sample
-- For each Sample in box_sample, create entries for all its SampleItems
INSERT INTO box_sample_item (
                id, shipping_box_id, sample_item_id, added_date,
                position_in_box, reception_status, reception_notes,
                sys_user_id, lastupdated
            )
            SELECT
                nextval('box_sample_item_seq'),
                bs.shipping_box_id,
                si.id as sample_item_id,
                bs.added_date,
                bs.position_in_box,
                bs.reception_status,
                bs.reception_notes,
                bs.sys_user_id,
                bs.lastupdated
            FROM box_sample bs
            JOIN sample_item si ON si.samp_id::text = bs.sample_id::text
            WHERE (si.rejected IS NULL OR si.rejected = false)
              AND (si.voided IS NULL OR si.voided = false)
            ORDER BY bs.id, si.id ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/shipment-012-migrate-to-sample-item.xml::2::pkomena
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
