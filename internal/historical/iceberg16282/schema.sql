CREATE TABLE IF NOT EXISTS historical_iceberg_runs (manifest_digest text PRIMARY KEY, complete boolean NOT NULL);
CREATE TABLE IF NOT EXISTS historical_iceberg_events (event_id text PRIMARY KEY, file_id text NOT NULL, event_offset integer NOT NULL UNIQUE);
CREATE TABLE IF NOT EXISTS historical_iceberg_offsets (singleton integer PRIMARY KEY CHECK (singleton=1), committed_offset integer NOT NULL);
CREATE TABLE IF NOT EXISTS historical_iceberg_offset_history (step integer PRIMARY KEY, committed_offset integer NOT NULL);
CREATE TABLE IF NOT EXISTS historical_iceberg_consumptions (step integer PRIMARY KEY, coordinator text NOT NULL, event_id text NOT NULL REFERENCES historical_iceberg_events(event_id), event_offset integer NOT NULL, file_id text NOT NULL);
CREATE TABLE IF NOT EXISTS historical_iceberg_snapshots (snapshot_id text PRIMARY KEY, coordinator text NOT NULL, step integer NOT NULL UNIQUE);
CREATE TABLE IF NOT EXISTS historical_iceberg_registrations (snapshot_id text PRIMARY KEY REFERENCES historical_iceberg_snapshots(snapshot_id), file_id text NOT NULL, event_id text NOT NULL REFERENCES historical_iceberg_events(event_id), coordinator text NOT NULL, step integer NOT NULL UNIQUE);
CREATE TABLE IF NOT EXISTS historical_iceberg_trace (step integer PRIMARY KEY, kind text NOT NULL, coordinator text NOT NULL, event_id text NOT NULL, file_id text NOT NULL, snapshot_id text NOT NULL, committed_offset integer NOT NULL);
