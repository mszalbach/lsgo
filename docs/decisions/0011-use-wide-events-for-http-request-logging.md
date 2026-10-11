# Use wide events for HTTP request logging

## Context and Problem Statement

For observability, the application should log events, especially errors.

## Considered Options

* Write separate log entries for each request step
* Collect request-scoped attributes in a wide event and emit them once when the request finishes

## Decision Outcome

Chosen option: "Collect structured log attributes in a request-scoped wide event", because this keeps related request data together, reduces log noise, and preserves diagnostic visibility for failures without flooding the log stream with every successful transaction.

### Consequences

* Good, because request and handler details are available together in a single structured record.
* Good, because logs remain easier to correlate when diagnosing behavior.
* Bad, because the implementation must ensure the event is always written regardless of panics or errors in the request.
* Bad, because request handlers must include useful diagnostic attributes for them to appear in the completion record.
