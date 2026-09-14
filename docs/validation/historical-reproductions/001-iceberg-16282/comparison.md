# Historical mechanism comparison

The machine-readable [result](result.json) enumerates public claims, local representations, PostgreSQL observations, match statuses, and limitations. Local comparison vocabulary is MATCH, PARTIAL, MISMATCH, UNKNOWN; it does not alter applicability or execution verdicts.

| Public claim | Local representation | Observed bounded state | Status | Limit |
| --- | --- | --- | --- | --- |
| Worker emits DATA_WRITTEN for X | One event identity | One PostgreSQL event row | MATCH | Synthetic control topic |
| C1 commits X in S1 | First snapshot registration | S1/file-X committed | MATCH | Reference table model |
| Offset gap before shutdown | Durable offset plus history | Offset 6 below event 7 at gap, shutdown, recovery | MATCH | No broker trace |
| C2 replays same event | Second consumption of same ID | Same event ID and file-X at step 9 | MATCH | Logical sequential actors |
| Same file appears in S1 and S2 | Two distinct snapshots and registrations | file-X in S1 and S2 | MATCH | Row-count symptom is not simulated |

**Local historical comparison: MATCH; milestone recommendation: PASS**, conditional on the checked PostgreSQL run. PASS means only that FailureMesh represented and deterministically reproduced the reported mechanism within this bounded local model. It does not independently prove the upstream bug or its proposed root cause. See [source record](source-record.md) for reported observations, analysis, speculation, unknowns, and the catalog conflict. No exact Iceberg 1.10.1 runtime, Kafka broker, GCS, REST/JDBC catalog, real coordinator scheduling, 20-minute race, upstream fix, general family-10 portability, fourth PFC, customer applicability, production readiness, or market validation is established. Expert reviews remain outstanding.
