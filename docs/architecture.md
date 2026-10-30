# 1. Context & Scope

```mermaid
C4Context
    title System Context View for LSGo
    Person(user, "User", "Uses a web browser to access files")
    System(lsgo, "LSGo", "A Go application that serves a folder as a web page")
    System_Ext(folder, "Folder", "A folder on the system where LSGo runs")

    Rel(user, lsgo, "Uses")
    Rel(lsgo, folder, "Lists folders and files")

    UpdateLayoutConfig($c4ShapeInRow="1", $c4BoundaryInRow="1")
```

# 2. Building Block View 

```mermaid
C4Component
    title Component View for LSGo

    Person(user, "User", "Uses a web browser to access files")

    Container_Boundary(lsgo, "LSGo") {

        Container(webPkg, "web", "Go + HTML", "Provides a web interface for browsing folders and viewing files")
        Container(assetsPkg, "assets", "Folder", "HTML templates, CSS, JavaScript, etc.")
        Container(cmdClient, "main", "Go", "Main entry point that initializes and coordinates the system")
        Container(filesystemPkg, "filesystem", "Go", "Provides functionality for listing and working with folder structures")

        Container_Boundary(filesystem, "Filesystem") {
            System_Ext(folder, "Folder", "A folder on the system where LSGo runs")
        }
    }

    Rel(user, webPkg, "Browses folders and views or downloads files")
    Rel(webPkg, assetsPkg, "Loads UI assets")
    Rel(webPkg, filesystemPkg, "Accesses the file system")

    Rel(filesystemPkg, folder, "Reads folders and files")

    UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="1")

    UpdateRelStyle(user, webPkg, $offsetY="-40", $offsetX="0")
    UpdateRelStyle(webPkg, assetsPkg, $offsetY="-20", $offsetX="-35")
    UpdateRelStyle(filesystemPkg, folder, $offsetY="-20", $offsetX="45")
```

## 3. Runtime View

```mermaid
flowchart LR
    browser[Browser]
    request[GET /files/PATH<br/>untrusted URL path]
    normalize[web<br/>normalize URL path]
    api[filesystem<br/>root-relative lookup]
    boundary{os.Root containment<br/>configured folder only}
    folder[web<br/>serve folder]
    response[HTTP response]
    filesystem[(Configured folder)]

    browser --> request
    request --> normalize
    normalize --> api
    api --> boundary
    boundary -->|lookup succeeds| folder
    boundary -->|missing or invalid path| response
    boundary -.-> filesystem
    folder --> response
    response --> browser

    classDef untrusted fill:#fff3cd,stroke:#b58105,color:#3d2f00
    classDef trusted fill:#d1e7dd,stroke:#28734f,color:#123d2a
    class request,folder,filesystem,browser untrusted
    class normalize,api,response trusted
```

# 4. Deployment View 

```mermaid
C4Deployment
    title Deployment View for LSGo with OAuth2 Proxy

    Person(user, "User", "Uses a web browser to access files")

    Deployment_Node(applicationHost, "Application host", "Server") {
        Container(oauth2Proxy, "oauth2-proxy", "Reverse proxy", "Authenticates requests and forwards authorized requests to LSGo")
        Container(lsgo, "LSGo", "Go application", "A Go application that serves a folder as a web page")
        System_Ext(folder, "Folder", "A folder on the system where LSGo runs")
    }

    Deployment_Node(identityProviderHost, "External identity provider", "Hosted OpenID Connect service") {
        System_Ext(identityProvider, "Identity Provider", "Authenticates users using OpenID Connect")
    }

    Rel(user, oauth2Proxy, "Accesses LSGo over HTTPS")
    Rel(oauth2Proxy, identityProvider, "Authenticates users using OIDC")
    Rel(oauth2Proxy, lsgo, "Forwards authorized requests over the private network")
    Rel(lsgo, folder, "Reads folders and files")

    UpdateLayoutConfig($c4ShapeInRow="1", $c4BoundaryInRow="2")

    UpdateRelStyle(user, oauth2Proxy, $offsetY="-40", $offsetX="0")
    UpdateRelStyle(oauth2Proxy, identityProvider, $offsetY="60", $offsetX="-70")
```

# 5. Architecture Decisions (ADRs)

The project keeps its ADRs under [docs/decisions](./decisions).

# 6. Risks and Technical Debt

Priority is rated Low / Medium / High.

| What                  | Type | Priority | Description                                                                                                   |
| :-------------------- | :--- | :------- | :------------------------------------------------------------------------------------------------------------ |
| `go-rod` is abandoned | Debt | Low      | Only used for the e2e tests for Browser tests.                                                                |
| GO-2026-5932          | Risk | Low      | `golang.org/x/crypto` is used by testcontainers. Only used for the e2e tests to manage the Docker containers. |