# Context

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

# Container

The main package is allowed to access the other packages so it can create and initialize them.
The Mermaid C4 diagrams lack styling, which can make links harder to read.

```mermaid
C4Container
    title Container View for LSGo

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

# Dynamic diagrams

## Listing folders

The web layer receives an untrusted path value from the request. It normalizes
that value and passes it to `filesystem`, which is the security boundary for
filesystem access. `filesystem` is created with `os.OpenRoot`, so lookups are
relative to the configured folder and cannot escape that root. Only after the
lookup succeeds does the web layer render a folder response.

This should be a C4 Dynamic diagram, but Mermaid support is not quite there yet.

The yellow blocks mark parts of the system where untrusted content may appear
and must be handled carefully, such as URL encoding or correcting MIME types
before content is served to a user.


```mermaid
flowchart LR
    browser[Browser]
    request[GET /files/PATH<br/>untrusted URL path]
    normalize[web<br/>prefix ./ and filepath.Clean]
    api[explorer<br/>root-relative lookup]
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
