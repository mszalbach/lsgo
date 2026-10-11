# End-to-end tests with `go-rod`

## Context and Problem Statement

Browser-based tests need a framework that can control a browser and expose a straightforward API.

## Considered Options

* Use `go-rod` for browser control
* Use another browser automation library, such as `chromedp` 
* Use a Node-based solution like `Playwright`

## Decision Outcome

Chosen option: "Use `go-rod` for the initial end-to-end browser tests", because it was the first suitable option found, its setup is lower-friction than lower-level tools, and it downloads a browser without requiring a separate language ecosystem.

### Consequences

* Good, because browser automation is simple to start with and works well for the project’s initial e2e coverage.
* Bad, because `go-rod` appears somewhat outdated and may need replacement later if maintenance lags.
