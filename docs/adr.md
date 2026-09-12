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


## 20260912-1 Absolute links instead of relative

accepted

### Context

The web interface contains links to files, directories, breadcrumbs, and static assets. Relative links would depend on the current directory URL and could make navigation harder to reason about, especially when paths contain special characters.

### Decision

Use absolute links for the web interface routes and static assets. File and directory links start at `/files`, while static assets start at `/static`.

### Consequences

- Links have a predictable target regardless of the current directory.
- Route construction is easier to understand and maintain.
- The application cannot work unchanged behind a reverse proxy that mounts it below an extra path segment, such as `/lsgo`; this requires adding a configurable base path concept.

## 20260912-2 Do not trust the files and folders

accepted

### Context

LSGo exposes a user-selected folder through a web interface. File and directory names originate from the filesystem and may contain characters that have a meaning in URLs or HTML. File content could contain JavaScript, or a file could be too large to render.

The application must also avoid allowing paths to escape the selected root folder.

### Decision

Treat file and directory names as untrusted input. Escape paths when creating URL values, use the standard HTML template escaping, and resolve filesystem access relative to an `os.Root` for the configured folder.

Treat file content as untrusted input. Only render files with safe MIME types or provide a download instead.

If files are too large, show a warning and offer a download link instead of trying to render them.

### Consequences

- File names are safely represented in generated links and HTML.
- Access is constrained to the configured root instead of the process working directory.
- Files and folders with unusual names require careful URL and filesystem handling and may expose edge cases that need dedicated tests.
- Special MIME type handling is required instead of using Go's default way of serving files.
