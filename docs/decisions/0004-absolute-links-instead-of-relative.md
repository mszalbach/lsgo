# Absolute links instead of relative

## Context and Problem Statement

The web interface contains links to files, folders, breadcrumbs, and static assets.

## Considered Options

* Use relative URLs for files, folders, and static assets
* Use absolute URLs rooted at the application base path

## Decision Outcome

Chosen option: "Use absolute links for the web interface routes and static assets", because absolute links are predictable regardless of the current folder context and remain easier to build and maintain. 
Relative links depend on the current URL and can become hard to reason about when navigation contexts change or paths contain special characters.

### Consequences

* Good, because links have a consistent target and route construction is clearer.
* Good, because browser navigation and generated URLs are easier to reason about.
* Bad, because the application cannot work unchanged behind an extra reverse-proxy prefix without adding a configurable base path.
