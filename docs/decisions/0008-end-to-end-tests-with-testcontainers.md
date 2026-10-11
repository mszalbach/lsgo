# End-to-end tests with Testcontainers

## Context and Problem Statement

End-to-end tests need a managed container that can be started, accessed, and stopped reliably in a repeatable manner.

## Considered Options

* Manually manage the container lifecycle in the test suite
* Use Testcontainers to manage the container used by end-to-end tests

## Decision Outcome

Chosen option: "Use Testcontainers to manage the containers used by end-to-end tests", because testcontainers provides a defined lifecycle and support for mounting files and binding random ports, which makes end-to-end test setup more reliable.

### Consequences

* Good, because container startup and teardown are handled in a consistent way.
* Good, because the tests can rely on random ports and mounted file systems.
* Bad, because the tests depend on a container runtime being available in the environment.
