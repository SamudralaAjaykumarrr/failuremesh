CREATE TABLE IF NOT EXISTS business_operations (
  run_id text NOT NULL,
  logical_id text NOT NULL,
  status text NOT NULL,
  PRIMARY KEY (run_id, logical_id)
);
CREATE TABLE IF NOT EXISTS provider_effects (
  effect_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  run_id text NOT NULL,
  logical_id text NOT NULL,
  attempt integer NOT NULL,
  committed_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX IF NOT EXISTS provider_effects_run_logical ON provider_effects (run_id, logical_id);
CREATE TABLE IF NOT EXISTS provider_idempotency (
  run_id text NOT NULL,
  logical_id text NOT NULL,
  effect_id bigint NOT NULL REFERENCES provider_effects(effect_id),
  PRIMARY KEY (run_id, logical_id)
);
