# Evidence and execution verdicts

Here **proof** means an auditable empirical execution argument within a declared scope; it is not formal verification or a mathematical proof. The verifier is deterministic over a versioned rule set and an evidence bundle. It must validate evidence completeness and integrity before interpreting outcomes.

An evidence bundle binds PFC ID/version, AC version and decisive facts, application/build identity, environment, invariant, experiment manifest/space, adapter and verifier versions, safety plan/authorization, run IDs, timestamps or causal ordering, stimuli, fault placement, observations, observer completeness, and artifact digests/provenance. Sensitive raw artifacts stay local. Evidence must distinguish a logical operation from each physical provider request and each committed provider effect. A digest proves artifact identity, not truth; trustworthy observations require an authoritative or independently corroborated source.

| Verdict | Exact meaning |
| --- | --- |
| `EXPOSED` | Valid evidence from an authorized run demonstrates the PFC's causal sequence and a violation of its declared invariant within the scoped build/environment. A reproducible counterexample is attached. |
| `PROVEN_RESILIENT` | The declared finite experiment space was executed under valid safety controls; all required observations are complete; the invariant held for every in-scope run. The result is scoped to PFC version, application/build identity, environment, invariant, experiment space, and collected evidence. It does not extrapolate to untested faults or future builds. |
| `UNKNOWN` | Neither claim is justified, including missing or conflicting evidence, incomplete space, failed observer, invalid/aborted run, unverified fault placement, or uncertain causal mapping. The reason is recorded. |

An exposure counterexample must contain a minimal replay recipe (pinned versions/build/configuration and initial state), logical operation ID, exact fault trigger and timing, ordered requests/retries/commits, invariant evaluation, expected versus observed outcome, raw local evidence references/digests, and reset instructions. Repeatable reproduction is required for promotion beyond L1; a single observed violation may still be `EXPOSED` if evidence is complete, with reproducibility status explicit. A resilience bundle includes the executed space, every run outcome, observer completeness, and exclusions. Empty or skipped experiment sets cannot yield `PROVEN_RESILIENT`.
