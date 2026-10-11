# Not using npm for UI dependencies

## Context and Problem Statement

The UI may need JavaScript and CSS frameworks.

## Considered Options

* Use npm and package.json to manage UI dependencies
* Use a CDN or manual asset management instead

## Decision Outcome

Chosen option: "Do not use npm for UI dependencies", because the simple UI does not need complex framework management, and avoiding npm reduces dependency churn and security concerns.

### Consequences

* Good, because the project avoids a large JavaScript dependency tree.
* Good, because this is easier to reason about and fits the lightweight nature of the interface.
* Bad, because uI dependencies and versions must be managed through another mechanism, possibly with custom tooling or Renovate support.
