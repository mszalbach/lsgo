# Zip files are directly written to `http.ResponseWriter`

## Context and Problem Statement

ZIP archives can be created in a temporary file before serving, or they can be streamed directly to the HTTP response.
## Considered Options

* Create a temporary archive file before sending it to the client
* Write the ZIP stream directly to the HTTP response

## Decision Outcome

Chosen option: "Write ZIP archives directly to the HTTP response", because it keeps the implementation simple, minimizes memory and disk usage, and lets the client begin downloading immediately.

### Consequences

* Good, because the implementation is straightforward and avoids temporary file handling.
* Good, because clients can start downloading before the archive is complete.
* Bad, because errors after the response begins and client disconnections cannot be handled cleanly.
* Bad, because the design may need to change for very large folders or cases where better error handling becomes critical.
