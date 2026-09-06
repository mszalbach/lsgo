# ADRs

## 20260906-1 Using Go

accepted

### Context

The author wants to improve their Go skills. Go is a commonly used language for command-line tools and Kubernetes-related tooling.

### Decision

This project will be implemented in Go.

### Consequences

- The codebase and development workflow will follow Go conventions (modules, build tooling, dependency management).
- Choosing Go influences CI, packaging, and contributor expectations.

## 20260906-2 Using `testify` for assertions

accepted

### Context

The project requires tests with clear assertions. The author is more familiar with assertion-style testing from other languages and prefers concise assertion helpers for readability.

### Decision

Use the `github.com/stretchr/testify/assert` assertion helpers to make tests more expressive and easier to read.

### Consequences

- Adds a testing dependency
- Tests may be easier to read and write, especially for those coming from other ecosystems
- This may reduce emphasis on the minimal standard-library testing style


## 20260906-3 Not using npm for UI dependencies

accepted

### Context

The UI may require JavaScript and CSS frameworks to look good without requiring too much web development.
One option would be to create a package.json and use npm to manage them. Other solutions are to use them directly via a CDN
or download them manually.

### Decision

npm will not be used because this ecosystem has many security issues and tends to download many dependencies.
The simple UI should not require complex frameworks.

### Consequences

- UI dependencies and versions must be managed by another mechanism.
- A custom mechanism that works with Renovate will probably be needed.
