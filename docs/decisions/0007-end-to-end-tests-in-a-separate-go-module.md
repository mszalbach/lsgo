# End-to-end tests in a separate Go module

## Context and Problem Statement

LSGo can benefit from end-to-end tests that verify the Docker container and browser-based navigation in a realistic environment. Such tests need extra dependencies and should not reuse production code for testing.

## Considered Options

* Keep end-to-end tests alongside the main Go tests
* Place end-to-end tests in a separate folder
* Store end-to-end tests in a separate Go module

## Decision Outcome

Chosen option: "Store the end-to-end tests in a separate Go module", because this keeps production dependencies clean and separates test dependencies from application code while preserving a clear testing boundary.

### Consequences

* Good, because the production `go.sum` stays smaller and test dependencies do not affect the app module.
* Good, because the tests remain separated from production code and can exercise realistic runtime behavior.
* Bad, because cI and local test execution must handle multiple Go modules.
