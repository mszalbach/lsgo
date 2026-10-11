# Do not trust the files and folders

## Context and Problem Statement

LSGo exposes a user-selected folder through a web interface. File and folder names come directly from the filesystem and may contain characters with special meaning in URLs or HTML, and file content may also contain untrusted script content or be too large to render safely. The application must also prevent path traversal outside the configured root folder.

## Considered Options

* Treat file names, paths, and content as trusted UI input
* Treat files, folders, and contents as untrusted input and validate or escape them

## Decision Outcome

Chosen option: "Treat file and folder names as untrusted input and constrain filesystem access to the configured root", because the application serves a user-chosen directory and must avoid path traversal, unsafe HTML rendering, and URL-breaking characters while still allowing normal browsing.

### Consequences

* Good, because generated links and HTML output are safe for filenames and user-controlled paths.
* Good, because access is limited to the configured root instead of the process working directory.
* Bad, because unusual file names and formats require careful URL handling and special MIME treatment.
