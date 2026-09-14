CREATE TABLE IF NOT EXISTS consumer_effects (
 effect_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 run_id text NOT NULL,
 message_id text NOT NULL,
 delivery_attempt integer NOT NULL CHECK (delivery_attempt IN (1, 2))
);
CREATE INDEX IF NOT EXISTS consumer_effects_run_message ON consumer_effects (run_id, message_id);
CREATE TABLE IF NOT EXISTS consumer_idempotency (
 run_id text NOT NULL,
 message_id text NOT NULL,
 effect_id bigint NOT NULL REFERENCES consumer_effects(effect_id),
 PRIMARY KEY (run_id, message_id)
);
