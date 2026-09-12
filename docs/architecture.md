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

The main package is allowed to access all other packages to create and initialize them.
The Mermaid C4 diagrams lack styling, which can lead to readability problems when links are present.

```mermaid
C4Container
    title Container View for LSGo

    Person(user, "User", "Uses a web browser to access files")

    Container_Boundary(lsgo, "LSGo") {

        Container(webPkg, "web", "Go + HTML", "Provides a web interface for browsing folders and viewing files")
        Container(assetsPkg, "assets", "Directory", "HTML templates, CSS, JavaScript, etc.")
        Container(cmdClient, "main", "Go", "Main entry point that initializes and coordinates the system")
        Container(explorerPkg, "explorer", "Go", "Provides functionality for listing and working with folder structures")

        Container_Boundary(filesystem, "Filesystem") {
            System_Ext(folder, "Folder", "A folder on the system where LSGo runs")
        }
    }

    Rel(user, webPkg, "Browses folders and views or downloads files")
    Rel(webPkg, assetsPkg, "Loads UI assets")
    Rel(webPkg, explorerPkg, "Accesses the file system")

    Rel(explorerPkg, folder, "Reads folders and files")

    UpdateLayoutConfig($c4ShapeInRow="2", $c4BoundaryInRow="1")

    UpdateRelStyle(user, webPkg, $offsetY="-40", $offsetX="0")
    UpdateRelStyle(webPkg, assetsPkg, $offsetY="-20", $offsetX="-35")
    UpdateRelStyle(explorerPkg, folder, $offsetY="-20", $offsetX="45")
```
