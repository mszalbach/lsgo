# Focus on integration tests instead of unit tests

## Context and Problem Statement

The main behavior of LSGo is exposed through its HTTP interface.

## Considered Options

* Prefer unit tests for internal implementation details
* Prefer integration tests for user-visible behavior and add unit tests only for narrow edge cases

## Decision Outcome

Chosen option: "Prefer integration tests for user-visible behavior and use them as the default testing style", because the most important behaviors are externally observable, and integration tests verify that the system works as users actually experience it.

Users interact with routes, rendered HTML, headers, links, and file contents rather than individual Go functions or types. Unit tests for implementation details can make refactorings harder.

### Consequences

* Good, because internal naming and architecture can change without broad test rewrites when behavior stays the same.
* Good, because integration tests exercise the real interaction between components.
* Bad, because tests may be slower and require more setup than isolated unit tests.
* Bad, because failures may be less localized and need clearer test names and assertions.
