-- source: liquibase liquibase/3.5.x.x/shipment-011-add-sample-shipment-referral-type.xml::1-add-sample-shipment-referral-type::pkomena
-- +goose Up
-- +goose StatementBegin
-- Add 'Sample Shipment' referral type for shipment module
INSERT INTO referral_type (id, name, description, display_key) VALUES (nextval('referral_type_seq'), 'Sample Shipment', 'Sample sent for testing via shipment', 'referral.type.sampleShipment') ON CONFLICT DO NOTHING;
INSERT INTO localization (id, description, lastupdated) VALUES (nextval('localization_seq'), 'referral type - sample shipment', NOW()) ON CONFLICT DO NOTHING;
INSERT INTO localization_value (id, localization_id, locale, value, last_updated) VALUES (nextval('localization_value_seq'), (SELECT id FROM clinlims.localization WHERE description = 'referral type - sample shipment'), 'en', 'Sample Shipment', NOW()) ON CONFLICT DO NOTHING;
INSERT INTO localization_value (id, localization_id, locale, value, last_updated) VALUES (nextval('localization_value_seq'), (SELECT id FROM clinlims.localization WHERE description = 'referral type - sample shipment'), 'fr', 'Envoi d''échantillon', NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/shipment-011-add-sample-shipment-referral-type.xml::1-add-sample-shipment-referral-type::pkomena
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
