CREATE TABLE IF NOT EXISTS lease_resources (
 resource_id text PRIMARY KEY,
 run_id text NOT NULL,
 current_owner text NOT NULL,
 current_token bigint NOT NULL CHECK (current_token > 0),
 logical_tick integer NOT NULL,
 expires_at_tick integer NOT NULL,
 lease_state text NOT NULL CHECK (lease_state IN ('active','expired'))
);
CREATE TABLE IF NOT EXISTS lease_history (
 run_id text NOT NULL,
 resource_id text NOT NULL,
 step integer NOT NULL,
 worker_id text NOT NULL,
 fencing_token bigint NOT NULL,
 action text NOT NULL CHECK (action IN ('acquired','expired')),
 logical_tick integer NOT NULL,
 PRIMARY KEY (run_id,step)
);
CREATE TABLE IF NOT EXISTS lease_trace (
 run_id text NOT NULL,
 seq integer NOT NULL,
 resource_id text NOT NULL,
 worker_id text NOT NULL,
 fencing_token bigint NOT NULL,
 kind text NOT NULL,
 PRIMARY KEY (run_id,seq)
);
CREATE TABLE IF NOT EXISTS protected_effects (
 run_id text NOT NULL,
 resource_id text NOT NULL,
 seq integer NOT NULL,
 worker_id text NOT NULL,
 fencing_token bigint NOT NULL,
 effect_value text NOT NULL,
 PRIMARY KEY (run_id,seq)
);
CREATE TABLE IF NOT EXISTS protected_attempts (
 run_id text NOT NULL,
 resource_id text NOT NULL,
 seq integer NOT NULL,
 worker_id text NOT NULL,
 fencing_token bigint NOT NULL,
 accepted boolean NOT NULL,
 PRIMARY KEY (run_id,seq)
);
