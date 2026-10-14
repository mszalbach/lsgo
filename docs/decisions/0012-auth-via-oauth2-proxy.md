# Use OAuth2 Proxy for authentication

## Context and Problem Statement

LSGo serves files that should not be accessible to everyone. Access to the application therefore needs to be protected by authentication.

## Considered Options

* Use [coreos/go-oidc](https://github.com/coreos/go-oidc/tree/v3) directly in LSGo
* Use [zitadel/oidc](https://github.com/zitadel/oidc) directly in LSGo
* Use [oauth2-proxy](https://github.com/oauth2-proxy/oauth2-proxy) in front of LSGo

## Decision Outcome

Chosen option: "Use `oauth2-proxy` in front of LSGo", because LSGo currently has no need to manage authentication or user details itself. `oauth2-proxy` also uses `coreos/go-oidc` internally.

### Consequences

* Good, because LSGo does not need to implement or maintain authentication flows and user management.
* Good, because authentication can be handled by a dedicated proxy rather than adding authentication-specific dependencies and responsibilities to LSGo.
* Bad, because deploying and configuring LSGo also requires deploying and configuring `oauth2-proxy` and its identity provider.
* Bad, because the deployment must ensure that requests cannot bypass the proxy and reach LSGo directly.
* Bad, because LSGo depends on `oauth2-proxy` for authentication behavior and must account for its security updates and compatibility.
