# Using `testify` for assertions

## Context and Problem Statement

The project needs readable and expressive tests. The author prefers assertion-style testing and wants concise helpers for verifying behavior without verbose boilerplate.

## Considered Options

* Use `github.com/stretchr/testify/assert` for assertions
* Use the Go standard library testing assertions and handwritten checks

## Decision Outcome

Chosen option: "Use `github.com/stretchr/testify/assert`", because it makes tests clearer and more readable, especially for contributors who are already accustomed to assertion-style testing patterns.

### Consequences

* Good, because test intent is easier to read and write.
* Good, because assertion helpers reduce repetitive verification code.
* Bad, because the project gains an additional dependency and relies on a style that is less minimal than the standard library alone.
